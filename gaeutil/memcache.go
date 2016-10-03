package gaeutil

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/memcache"
)

const Chunksize = 950000  // A single memcache item can't be bigger than 1000000 bytes
const MaxChunks = 32      // We only shard this much.

func singletonMCKey(name string) string { return "singleton:"+name }

func DeleteSingletonFromMemcache(c context.Context, name string) (error) {
	return memcache.Delete(c, name)
}

func LoadSingletonFromMemcache(c context.Context, name string) ([]byte, error) {
	item,err := memcache.Get(c, singletonMCKey(name))
	if err != nil {
    return nil, err  // might be memcache.ErrCacheMiss
	}
	return item.Value, nil
}
	
func SaveSingletonToMemcache(c context.Context, name string, data []byte) error {
	if len(data) > 950000 {
		return fmt.Errorf("singleton too large (name=%s, size=%d)", name, len(data))
	}
	item := &memcache.Item{Key:singletonMCKey(name), Value:data}
	return memcache.Set(c, item)
}

func LoadShardedSingletonFromMemcache(c context.Context, name string) ([]byte, error) {
	return bytesFromMemcacheShards(c, name)
}
func SaveShardedSingletonToMemcache(c context.Context, name string, data []byte) error {
	return bytesToMemcacheShards(c, name, data)
}


func bytesToMemcacheShards(c context.Context, key string, b []byte) error {
	if len(b) > MaxChunks * Chunksize {
		return fmt.Errorf("obj '%s' was too large; %d > %d", key, len(b), MaxChunks * Chunksize)
	}

	items := []*memcache.Item{}
	for i:=0; i<len(b); i+=Chunksize {
		k := fmt.Sprintf("=%d=%s",i,key)
		s,e := i, i+Chunksize-1
		if e>=len(b) { e = len(b)-1 }
		items = append(items, &memcache.Item{ Key:k , Value:b[s:e+1] }) // slice sytax is [s,e)
	}

	return memcache.SetMulti(c, items)
}

// err might be memcache.ErrCacheMiss
func bytesFromMemcacheShards(c context.Context, key string) ([]byte, error) {
	keys := []string{}
	for i:=0; i<32; i++ { keys = append(keys, fmt.Sprintf("=%d=%s",i*Chunksize,key)) }

	if items,err := memcache.GetMulti(c, keys); err != nil {
		return nil, fmt.Errorf("MCShards/GetMulti/'%s' err: %v\n", key, err)

	} else {
		b := []byte{}
		for i:=0; i<32; i++ {
			if item,exists := items[keys[i]]; exists==false {
				break
			} else {
				b = append(b, item.Value...)
			}
		}

		if len(b) > 0 {
			return b, nil
		} else {
			return nil, memcache.ErrCacheMiss
		}
	}
}



