package gaeutil

import(
	"fmt"
	
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)


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
