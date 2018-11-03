package singleton

// This is an implementation of the singleton interface.
// It requires a datastore provider; singletons are read/written to datastore.
// Implement the singleton interface, on top of the datastore interface.

/*

import "github.com/skypies/util/gcp/ds"
import "github.com/skypies/util/gcp/singleton"

p := ds.NewCloudProvider(...)
s := singleton.NewProvider(p)

*/

import(
	"bytes"
	"encoding/gob"

	"golang.org/x/net/context"

	"github.com/skypies/util/gcp/ds"

	"github.com/skypies/util/singleton"
)

type SingletonProvider struct {
	ds.DatastoreProvider
	ErrIfNotFound bool
}

func NewProvider(p ds.DatastoreProvider) SingletonProvider {
	return SingletonProvider{p,false}
}

func (sp SingletonProvider)singletonDSKey(c context.Context, name string) ds.Keyer {
	return sp.NewNameKey(c, "Singleton", name, nil)
}

func (sp SingletonProvider)ReadSingleton(ctx context.Context, name string, obj interface{}) error {
	s := singleton.Singleton{}

	if err := sp.Get(ctx, sp.singletonDSKey(ctx,name), &s); err != nil {
		if err != ds.ErrNoSuchEntity {
			return err
		}

		// Some consumers need to know about this; some don't care.
		if sp.ErrIfNotFound {
			return err
		} else {
			sp.Warningf(ctx, "ReadSingleton('%s'), but ds.ErrNoSuchEntity; initializing ?", name)
		}
	}

	if s.Value == nil {
		// This happens if the object was not found; don't try to decode it.
		return nil
	}

	buf := bytes.NewBuffer(s.Value)

	if err := gob.NewDecoder(buf).Decode(obj); err != nil {
		return err
	}

	return nil
}


func (sp SingletonProvider)WriteSingleton(ctx context.Context, name string, obj interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(obj); err != nil {
		return err
	}

	data := buf.Bytes()

	if len(data) > 950000 {
		return singleton.ErrSingletonTooBig
	}

	s := singleton.Singleton{data}
	_,err := sp.Put(ctx, sp.singletonDSKey(ctx,name), &s)
	return err
}
