package singleton

import(
	"errors"

	"golang.org/x/net/context"
)

/*

type Foo struct {
  S string
}

sp := some.SingletonProvider{}

foo := Foo{S:"hello"}
err1 := sp.WriteSingleton(ctx, "Foo_007", &foo)

foo2 := Foo{}
err2 := sp.ReadSingleton(ctx, "Foo_007", &foo2)

*/

var(
	ErrNoSuchEntity = errors.New("util/singleton: no such entity")
	ErrSingletonTooBig = errors.New("util/singleton: object too big to write")
)

type Singleton struct {
	Value []byte `datastore:",noindex"`
}

type SingletonProvider interface {
	ReadSingleton(ctx context.Context, name string, obj interface{}) error
	WriteSingleton(ctx context.Context, name string, obj interface{}) error
}
