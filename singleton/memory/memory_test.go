package memory

// go test -v github.com/skypies/util/singleton/memory

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/skypies/util/singleton"
)

type Foo struct {
	S string
}

var ctx = context.Background()

func TestHappyPath(t *testing.T) {
	str := "Splendid. "
	name := "mem_singleton_name"

	foo1 := Foo{S:str}
	foo2 := Foo{}

	p := NewProvider()

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		t.Errorf("Memory Read noexist, err not a miss: %v", err)
	}

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Memory Write, err: %v", err)
	}

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Memory Read exist, err: %v", err)
	}
}

func TestFails(t *testing.T) {
	str := "Splendid. "
	name := "mem_singleton_name"

	foo1 := Foo{S:str}
	foo2 := Foo{}

	p := NewProvider()

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Memory Write, err: %v", err)
	}

	// Prove it should be returnable
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Memory Read exist, err: %v", err)
	}

	p.AlwaysFail = true
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		t.Errorf("Memory ReadFails should have failed, err: %v", err)
	}

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err == nil {
		t.Errorf("Memory WriteFails, did not fail")
	}

}
