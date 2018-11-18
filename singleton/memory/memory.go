package memory

// A very basic in-memory singleton provider, for other testing.

import(
	"fmt"
	"reflect"
	"golang.org/x/net/context"
	"github.com/skypies/util/singleton"
)

type MemorySingletonProvider struct {
	// This map stores pointers, and we follow those pointers during read operations
	Map          map[string]interface{}
	AlwaysFail   bool
}

func NewProvider() MemorySingletonProvider {
	return MemorySingletonProvider{Map:map[string]interface{}{}}
}

func (sp MemorySingletonProvider)ReadSingleton(ctx context.Context, name string, f singleton.NewReaderFunc, ptr interface{}) error {
	if sp.AlwaysFail {
		return singleton.ErrNoSuchEntity
	}

	orig,exists := sp.Map[name]
	if !exists {
		return singleton.ErrNoSuchEntity
	}

	// Assert that 'ptr' is indeed a pointer to something (and is not nil)
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("interface arg was '%s', expected pointer", reflect.TypeOf(ptr))
	}

	//fmt.Printf("Base Read\n orig/src = %T\n ptr/dest = %T\n[%v / %v]\n", orig, ptr,
	//	reflect.TypeOf(orig).Kind(), reflect.TypeOf(ptr).Kind())
	
	// Now copy whatever the map interface{} points to
	srcValue := reflect.ValueOf(orig).Elem() // Follows the pointer
	dstValue := reflect.ValueOf(ptr).Elem() // Follows the pointer
	dstValue.Set(srcValue)

	return nil
}

func (sp MemorySingletonProvider)WriteSingleton(ctx context.Context, name string, f singleton.NewWriteCloserFunc, ptr interface{}) error {
	if sp.AlwaysFail {
		return fmt.Errorf("MemorySingletonProvider asked to always fail")
	}

	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("obj to store was '%s', expected pointer", reflect.TypeOf(ptr))
	}

	//fmt.Printf("WaheyWrite\n ptr/src = %T\n", reflect.TypeOf(ptr).Kind())
	
	sp.Map[name] = ptr
	return nil
}
