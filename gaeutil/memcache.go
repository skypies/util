package gaeutil

import (
	"bytes"
	//"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
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

/*

 object := SomeThing{}

 gaeutil.SaveToMemcacheShards(ctx, "mything", &object)

 if err := gaeutil.LoadFromMemcacheShards(ctx, "mything", &object); err == nil {
    // use object
 }

 */

func SaveToMemcacheShards(ctx context.Context, name string, ptr interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(ptr); err != nil {
		return err
	}
	// log.Infof(ctx, "SAVE<%T>, %d bytes", ptr, buf.Len())
	
	return bytesToMemcacheShards(ctx, name, buf.Bytes())
}

func LoadFromMemcacheShards(ctx context.Context, name string, ptr interface{}) error {
	myBytes,err := bytesFromMemcacheShards(ctx, name)
	if err != nil { return err }
	
	buf := bytes.NewBuffer(myBytes)
	if err := gob.NewDecoder(buf).Decode(ptr); err != nil {
		log.Errorf(ctx, "LoadFromMemcacheShards err %v", err)
		return err
	}

	//log.Infof(ctx, "LOAD<%T>, %d/%d bytes", ptr, len(myBytes), buf.Len())

	return nil
}

/* TTL versions; if memcache obj older than 'duration', returns memcache.ErrCacheMiss

 object := SomeThing{}
 duration := time.Minute

 err := gaeutil.SaveToMemcacheShardsTTL(ctx, "mything", &object, duration)
 err := gaeutil.LoadFromMemcacheShardsTTL(ctx, "mything", &object)

 */

// Note that the pointer gets flattened in this roundtrip through encoding/gob.
// When submitted to Save*, the interface is a pointer to an object; but when retrived via
// Load*, the interface contains the actual object originally pointed to.
// There is some reflection nonsense needed to copy results between interfaces.
// https://github.com/mohae/deepcopy/blob/master/deepcopy.go
type explodingObj struct {
	Obj interface{}
	Expires time.Time
}

func SaveToMemcacheShardsTTL(ctx context.Context, name string, ptr interface{}, d time.Duration) error {
	return SaveToMemcacheShards(ctx, name, explodingObj{ptr, time.Now().Add(d)})
}

func LoadFromMemcacheShardsTTL(ctx context.Context, name string, ptr interface{}) error {
	eObj := explodingObj{}
	if err := LoadFromMemcacheShards(ctx, name, &eObj); err != nil {
		return err
	} else if time.Now().After(eObj.Expires) {
		return memcache.ErrCacheMiss
	}

	// Assert that 'ptr' is indeed a pointer to something (and is not nil)
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("interface arg was '%s', expected pointer", reflect.TypeOf(ptr))

	}
	if reflect.ValueOf(ptr).IsNil() {
		return fmt.Errorf("interface arg was nil pointer")
	}

	// Assert that ptr points to the same type of thing that we have in the explodey obj
	if reflect.ValueOf(ptr).Elem().Type() != reflect.TypeOf(eObj.Obj) {
		return fmt.Errorf("type mismatch; asked to load '%s' into '%s'",
			reflect.TypeOf(eObj.Obj), reflect.ValueOf(ptr).Elem().Type())
	}

	// Now copy whatever is inside of eObj.Obj into whatever the pointer points to
	srcValue := reflect.ValueOf(eObj.Obj)
	dstValue := reflect.ValueOf(ptr).Elem() // Follow the pointer
	dstValue.Set(srcValue)

	return nil
}


// The SaveSingletonToMemcache* routines are a kludge to allow appengine flexible
// environment apps (e.g. the pi/cmd/consolidator) to write into memcache, in the absence
// of any cloud APIs for doing so. The actual write is done by a friendly appengine standard
// app (e.g. flightdb/app/frontend), running a handler, which the appengine flexible app sends
// a POST to.

// memcacheSingletonEntry is for data transfer by HTTP
type memcacheSingletonEntry struct {
	Name   string
	Body []byte
}

// SaveSingletonToMemcacheHandler is a HTTP handler that expects an application/json request
// body that encodes a memcacheSingletonEntry struct. (The golang json libs take care of all
// the base64 encoding/decoding transparently). The entry will be written into memcache.
func SaveSingletonToMemcacheHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a JSON encoded request body", http.StatusBadRequest)
		return
	}

	entry := memcacheSingletonEntry{}
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, fmt.Sprintf("SSTMH/JsonDecode err:%v", err.Error()), http.StatusBadRequest)
		return
	} else if entry.Name == "" {
		http.Error(w, "SSTMH needs non-empty 'Name'", http.StatusBadRequest)
		return
	}

	if err := SaveSingletonToMemcache(ctx, entry.Name, entry.Body); err != nil {
		http.Error(w, fmt.Sprintf("SSTMH/memcache save err:%v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OK\n")
}

// SaveSingletonToMemcacheURL sends a HTTP request to the given URL, which should route to
// the SaveSingletonToMemcacheHandler routine above.
func SaveSingletonToMemcacheURL(name string, body []byte, url string) error {
	byt,_ := json.Marshal(memcacheSingletonEntry{Name:name, Body:body})

	resp,err := http.Post("http://"+url, "application/json; charset=utf-8", bytes.NewBuffer(byt))
	if err != nil {
		return fmt.Errorf("client.Post to %s err: %v\n", url, err)
	} else if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body,_ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s returned %q\n%s\n", url, resp.Status, body)
	}
	
	return nil
}
