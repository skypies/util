package memcache

// go test -v github.com/skypies/util/singleton/memcache

// This suite assumes a memcached on 127.0.0.1:11211

import (
	"encoding/gob"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/skypies/util/singleton"
)

type Foo struct {
	S string
}

func init() {
	gob.Register(Foo{})
}

var (
	ctx = context.Background()
	memcached = "127.0.0.1:11211"
)

func TestHappyPath(t *testing.T) {
	str := "Splendid. "
	name := "mc_singleton_name"

	foo1 := Foo{S:str}
	foo2 := Foo{}

	p := NewProvider(memcached)
	p.ErrIfNotFound = true

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		// This test may fail if the object happens to be present; no delete semantics yet !
		t.Errorf("Memcache Read noexist, err not a miss: %v\n", err)
	}

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Memcache Write, err: %v\n", err)
	}
	
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Memcache Read exist, err: %v\n", err)
	}
}

func TestSharding(t *testing.T) {
	str := strings.Repeat("Splendid. ", 1300000)
	name := "m2_singleton_name_2"

	foo1 := Foo{S:str}
	foo2 := Foo{}

	p := NewProvider(memcached)
	p.Timeout = time.Second * 4  // can take a while to read/write these big objects
	p.ShardCount = 32

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Memcache Sharded Write, err: %v\n", err)
	}
	
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Memcache Sharded Read, err: %v\n", err)
	} else if len(foo2.S) != len(foo1.S) {
		t.Errorf("Memcache Sharded Read, bad data: %d, %d\n", len(foo2.S), len(foo1.S))
	}
}
