package singleton

import(
	"compress/gzip"
	"errors"
	"io"

	"golang.org/x/net/context"
)

/*

type Foo struct {
  S string
}

sp := some.SingletonProvider{}

foo := Foo{S:"hello"}
foo2 := Foo{}

err1 := sp.WriteSingleton(ctx, "Foo_007", nil, &foo)
err2 := sp.ReadSingleton (ctx, "Foo_007", nil, &foo2)


foo3 := Foo{S:"hello, I will be gzipped"}
foo4 := Foo{}

err3 := sp.WriteSingleton(ctx, "Foo_007", GzipWriter, &foo3)
err4 := sp.ReadSingleton (ctx, "Foo_007", GzipReader, &foo4)

*/

var(
	ErrNoSuchEntity = errors.New("util/singleton: no such entity")
	ErrSingletonTooBig = errors.New("util/singleton: object too big to write")
)

type Singleton struct {
	Value []byte `datastore:",noindex"`
}

// These types might already exist in pkg/io ?
type NewReaderFunc func(io.Reader) (io.Reader, error)
type NewWriterFunc func(io.Writer) io.Writer

type SingletonProvider interface {
	// Functions can be nil; there are some gzip ones below.
	ReadSingleton (ctx context.Context, name string, f NewReaderFunc, ptr interface{}) error
	WriteSingleton(ctx context.Context, name string, f NewWriterFunc, ptr interface{}) error
}


// These two wrapper functions seem sadly needed, to launder method signature types
func GzipReader(rdr io.Reader) (io.Reader, error) {
	rdr,err := gzip.NewReader(rdr)
	return rdr,err
}

func GzipWriter(wtr io.Writer) io.Writer {
	return gzip.NewWriter(wtr)
}
