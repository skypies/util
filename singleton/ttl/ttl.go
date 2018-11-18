package ttl

// This package implements the Singleton interface using and applies TTL semantics. It relies
// on a user-provided SingletonProvider to do the read/writing.

import(
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"

	"github.com/skypies/util/singleton"
)

type TTLSingletonProvider struct {
	singleton.SingletonProvider
	TTL time.Duration // Objects older than this are transparently dropped
}

func NewProvider(d time.Duration, p singleton.SingletonProvider) TTLSingletonProvider {
	sp := TTLSingletonProvider {
		SingletonProvider: p,
		TTL: d,
	}

	return sp
}


// Note that the pointer gets flattened in this roundtrip through encoding/gob.
// When submitted to Save*, the interface is a pointer to an object; but when retrived via
// Load*, the interface contains the actual object originally pointed to.
// There is some reflection nonsense needed to copy results between interfaces.
// https://github.com/mohae/deepcopy/blob/master/deepcopy.go
type explodingObj struct {
	Obj interface{}
	Expires time.Time
}

func (sp TTLSingletonProvider)ReadSingleton(ctx context.Context, name string, f singleton.NewReaderFunc, ptr interface{}) error {
	eObj := explodingObj{}
	if err := sp.SingletonProvider.ReadSingleton(ctx, name, f, &eObj); err != nil {
		return err
	} else if time.Now().After(eObj.Expires) {
		return singleton.ErrNoSuchEntity
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

func (sp TTLSingletonProvider)WriteSingleton(ctx context.Context, name string, f singleton.NewWriteCloserFunc, ptr interface{}) error {
	o := explodingObj{ptr, time.Now().Add(sp.TTL)}
	return sp.SingletonProvider.WriteSingleton(ctx, name, f, o)
}
