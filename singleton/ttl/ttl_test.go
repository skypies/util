package ttl

// go test -v github.com/skypies/util/singleton/ttl

// This suite assumes memcache singletons work, and a memcached on 127.0.0.1:11211

import (
	"encoding/gob"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/skypies/util/singleton"
	"github.com/skypies/util/singleton/memcache"
)

type Foo struct {
	S string
}

var (
	ctx = context.Background()
	memcached = "127.0.0.1:11211"
	ttl = time.Second * 2
)

func init() {
	gob.Register(Foo{})
}

func TestHappyPath(t *testing.T) {
	str := "Splendid. "
	name := "singleton_name"

	foo1 := Foo{S:str}
	foo2 := Foo{}

	p := NewProvider(ttl, memcache.NewProvider(memcached))

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Write TTL, err: %v", err)
		return
	}
	
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Read unexpired TTL, err: %v", err)
	}

	time.Sleep(ttl + time.Second * 2)

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		t.Errorf("Read expired TTL, not a cache miss: %v", err)
	}
}
