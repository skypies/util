package memcache

// This package implements the util/singleton interface, on top of a memcache client lib.

import(
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	mclib "github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/net/context"

	"github.com/skypies/util/singleton"
)

const Chunksize = 950000  // A single memcache item can't be bigger than 1000000 bytes

type SingletonProvider struct {
	*mclib.Client
	ErrIfNotFound bool  // By default, we swallow 'item not found' errors.
	ShardCount    int   // defaults to no sharding. Set this if you need to store >Chunksize
}

func NewProvider(servers ...string) SingletonProvider {
	sp := SingletonProvider{
		Client: mclib.New(servers...),
	}

	return sp
}

func (sp SingletonProvider)ReadSingleton(ctx context.Context, name string, f singleton.NewReaderFunc, obj interface{}) error {
	var myBytes []byte
	var err error
	
	if sp.ShardCount > 2 {
		myBytes,err = sp.loadSingletonShardedBytes(name)
	} else {
		myBytes,err = sp.loadSingletonBytes(name)
	}

	if err == singleton.ErrNoSuchEntity {
		return err // Don't decorate this error
	} else if err != nil {
		return fmt.Errorf("ReadSingleton/loadBytes: %v", err)
	} else if myBytes == nil {
		// This happens if the object was not found; don't try to decode it.
		return nil
	}
	
	var reader io.Reader
	buf := bytes.NewBuffer(myBytes)

	reader = buf
	if f != nil {
		if reader,err = f(buf); err != nil {
			return fmt.Errorf("ReadSingleton/NewReaderFunc: %v", err)
		}
	}

	if err := gob.NewDecoder(reader).Decode(obj); err != nil {
		return fmt.Errorf("ReadSingleton/Decode: %v (%d bytes)", err, len(myBytes))
	}
	// fmt.Printf("(loaded %d bytes from memcache)\n", len(myBytes))
	
	return nil

}

func (sp SingletonProvider)WriteSingleton(ctx context.Context, name string, f singleton.NewWriteCloserFunc, obj interface{}) error {
	var buf bytes.Buffer
	var writer io.Writer
	var writecloser io.WriteCloser
	
	// All this type chicanery is so we can call Close() on the NewWriterFunc's thing, which is
	// needed for gzip.
	writer = &buf
	if f != nil {
		writecloser = f(&buf)
		writer = writecloser
	}

	if err := gob.NewEncoder(writer).Encode(obj); err != nil {
		return err
	}

	if f != nil {
		if err := writecloser.Close(); err != nil {
			return err
		}
	}

	data := buf.Bytes()

	if sp.ShardCount > 1 {
		return sp.saveSingletonShardedBytes(name, data)
	} else {
		return sp.saveSingletonBytes(name, data)
	}
}



func singletonMCKey(name string) string { return "singleton:"+name }

func (sp SingletonProvider)deleteSingleton(name string) (error) {
	return sp.Client.Delete(name)
}

func (sp SingletonProvider)loadSingletonBytes(name string) ([]byte, error) {
	item,err := sp.Client.Get(singletonMCKey(name))
	if err == mclib.ErrCacheMiss {
		// Swallow this error, if we need to.
		if sp.ErrIfNotFound {
			return nil, singleton.ErrNoSuchEntity
		}
		return nil, nil

	} else if err != nil {
		return nil, err
	}

	return item.Value, nil
}

func (sp SingletonProvider)saveSingletonBytes(name string, data []byte) error {
	if len(data) > Chunksize {
		return fmt.Errorf("singleton too large (name=%s, size=%d)", name, len(data))
	}
	item := mclib.Item{Key:singletonMCKey(name), Value:data}
	return sp.Client.Set(&item)
}

func (sp SingletonProvider)saveSingletonShardedBytes(key string, b []byte) error {
	if sp.ShardCount < 2 { return fmt.Errorf("saveSingletonShardedBytes: .ShardCount not set") }

	if len(b) > sp.ShardCount * Chunksize {
		return fmt.Errorf("obj '%s' was too large; %d > (%d shards x %d bytes)", key,
			len(b), sp.ShardCount, Chunksize)
	}

	// fmt.Printf("(saving over %d shards)\n", sp.ShardCount)
	
	for i:=0; i<len(b); i+=Chunksize {
		k := fmt.Sprintf("=%d=%s",i,key)
		s,e := i, i+Chunksize-1
		if e>=len(b) { e = len(b)-1 }

		item := mclib.Item{ Key:k , Value:b[s:e+1] } // slice sytax is [s,e)

		// fmt.Printf("  (saving shard %d ...)\n", i)
		if err := sp.Client.Set(&item); err != nil {
			return err
		}
		// fmt.Printf("  (... shard %d saved !)\n", i)
	}

	return nil
}

// err might be .ErrCacheMiss
func  (sp SingletonProvider)loadSingletonShardedBytes(key string) ([]byte, error) {
	if sp.ShardCount < 2 { return nil, fmt.Errorf("loadSingletonShardedBytes: .ShardCount not set") }
	
	keys := []string{}
	for i:=0; i<sp.ShardCount; i++ { keys = append(keys, fmt.Sprintf("=%d=%s",i*Chunksize,key)) }

	// fmt.Printf("(loading over %d shards)\n", sp.ShardCount)

	if items,err := sp.Client.GetMulti(keys); err != nil {
		return nil, fmt.Errorf("MCShards/GetMulti/'%s' err: %v\n", key, err)

	} else {
		b := []byte{}
		for i:=0; i<sp.ShardCount; i++ {
			if item,exists := items[keys[i]]; exists==false {
				break
			} else {
				// fmt.Printf("  (shard %d loaded)\n", i)
				b = append(b, item.Value...)
			}
		}

		if len(b) > 0 {
			return b, nil
		} else {
			return nil, singleton.ErrNoSuchEntity
		}
	}
}
