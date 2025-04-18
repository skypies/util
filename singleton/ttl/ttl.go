package ttl

// This package implements the Singleton interface using and applies TTL semantics. It relies
// on a user-provided SingletonProvider to do the read/writing.

import(
	"fmt"
	"reflect"
	"time"

	"context"

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

// Note: if the Obj interface is a pointer, and the whole object is roundtripped
// through encoding/gob, then it will eventually reappear flattened into a struct (gob
// does not attempt to recreate the indirection via pointer).
// So there is some reflection nonsense needed to copy results between interfaces, and to
// handle whether it has been flattened
// https://github.com/mohae/deepcopy/blob/master/deepcopy.go
type explodingObj struct {
	Obj interface{} // should be a pointer !
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

	// The extracted obj might still be a pointer (i.e. memory singleton), or if it went through
	// gob serialization, it might be flatted into an object. Handle both cases.
	srcValue := reflect.ValueOf(eObj.Obj)
	if srcValue.Type().Kind() == reflect.Ptr {
		srcValue = srcValue.Elem() // Follow the pointer
	}
	dstValue := reflect.ValueOf(ptr).Elem() // Follow the pointer
	
	// Assert that ptr points to the same type of thing that we have in the explodey obj
	if srcValue.Type() != dstValue.Type() {
		return fmt.Errorf("type mismatch; asked to load '%s' into '%s'",
			srcValue.Type(), dstValue.Type())
	}

	// Now copy whatever is inside of eObj.Obj into whatever the pointer points to
	dstValue.Set(srcValue)

	return nil
}

func (sp TTLSingletonProvider)WriteSingleton(ctx context.Context, name string, f singleton.NewWriteCloserFunc, ptr interface{}) error {
	o := explodingObj{ptr, time.Now().Add(sp.TTL)}
	return sp.SingletonProvider.WriteSingleton(ctx, name, f, &o)
}
