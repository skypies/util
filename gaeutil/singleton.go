package gaeutil

import(
	"bytes"
	"encoding/gob"
	"fmt"
	
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

/*

type Foo struct {
  S string
}

foo := Foo{S:"hello"}
err1 := gaeutil.SaveAnySingleton(ctx, "Foo_007", &foo)

foo2 := Foo{}
err2 := gaeutil.LoadAnySingleton(ctx, "Foo_007", &foo2)

 */

func singletonDSKey(c context.Context, name string) *datastore.Key {
	return datastore.NewKey(c, "Singleton", name, 0, nil)
}

type Singleton struct {
	Value []byte `datastore:",noindex"`
}

func LoadSingletonFromDatastore(c context.Context, name string) ([]byte, error) {
	s := Singleton{}
	if err := datastore.Get(c, singletonDSKey(c,name), &s); err != nil {
		return nil,err  // might be datastore.ErrNoSuchEntity
	}
	return s.Value,nil
}

func SaveSingletonToDatastore(c context.Context, name string, data []byte) error {
	if len(data) > 950000 {
		return fmt.Errorf("singleton too large (name=%s, size=%d)", name, len(data))
	}
	s := Singleton{data}
	_,err := datastore.Put(c, singletonDSKey(c,name), &s)
	return err
}

func LoadSingleton(c context.Context, name string) ([]byte, error) {
	data,err := LoadSingletonFromMemcache(c,name)
	if err != nil {
		// We don't care if it was a cache miss or a deeper error - failback to datastore either way
		data,err = LoadSingletonFromDatastore(c,name)
		if err == datastore.ErrNoSuchEntity {
			return nil,nil
		}
	}
	return data,nil
}

func SaveSingleton(c context.Context, name string, data []byte) error {
	if err := SaveSingletonToDatastore(c,name,data); err != nil {
		return err
	}

	SaveSingletonToMemcache(c,name,data)  // Don't care if this fails
	return nil
}


func LoadAnySingleton(ctx context.Context, name string, obj interface{}) error {
	myBytes,err := LoadSingleton(ctx, name)

	if err == datastore.ErrNoSuchEntity {
		// Strictly speaking, only LoadSingletonFromDatastore should expose this miss
		return nil
	} else if err != nil {
		return err
	}

	buf := bytes.NewBuffer(myBytes)

	if err := gob.NewDecoder(buf).Decode(obj); err != nil {
		return err
	}

	return nil
}

func SaveAnySingleton(ctx context.Context, name string, obj interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(obj); err != nil {
		return err
	}
	
	return SaveSingleton(ctx, name, buf.Bytes())
}
