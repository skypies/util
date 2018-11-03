package singleton

import(
	"errors"

	"golang.org/x/net/context"
)

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
