package combo

// go test -v github.com/skypies/util/singleton/combo

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/skypies/util/singleton"
	"github.com/skypies/util/singleton/memory"
)

type Foo struct {
	S string
}

var (
	ctx = context.Background()
	str = "Splendid. "
	name = "mem_singleton_name"
)

func TestHappyPath(t *testing.T) {
	foo1 := Foo{S:str}
	foo2 := Foo{}

	p1 := memory.NewProvider()
	p2 := memory.NewProvider()
	p := NewProvider(p1, p2)

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		t.Errorf("Memory Read noexist, err not a miss: %v\n", err)
	}

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Memory Write, err: %v\n", err)
	}
	
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Memory Read exist, err: %v\n", err)
	}
}


func TestFailingPrimary(t *testing.T) {
	foo1 := Foo{S:str}
	foo2 := Foo{}

	p1 := memory.NewProvider()
	p2 := memory.NewProvider()
	p1.AlwaysFail = true
	p := NewProvider(p1, p2)

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		t.Errorf("Memory Read noexist, err not a miss: %v\n", err)
	}

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err != nil {
		t.Errorf("Memory Write, err: %v\n", err)
	}
	
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != nil {
		t.Errorf("Memory Read exist, err: %v\n", err)
	}
}

func TestFailingSecondary(t *testing.T) {
	foo1 := Foo{S:str}
	foo2 := Foo{}

	p1 := memory.NewProvider()
	p2 := memory.NewProvider()
	p2.AlwaysFail = true
	p := NewProvider(p1, p2)

	if err := p.ReadSingleton(ctx, name, nil, &foo2); err != singleton.ErrNoSuchEntity {
		t.Errorf("Memory Read noexist, err not a miss: %v", err)
	}

	if err := p.WriteSingleton(ctx, name, nil, &foo1); err == nil {
		t.Errorf("Memory Write failing secondary, didn't fail")
	}

	// Now write the object
	p2.AlwaysFail = false
	if err := p.WriteSingleton(ctx, name, nil, &foo1); err == nil {
		t.Errorf("hmm, this write should have worked: %v", err)
	}

	// Now see the secondary read fail
	p2.AlwaysFail = true
	if err := p.ReadSingleton(ctx, name, nil, &foo2); err == nil {
		t.Errorf("Memory Read failing secondary, did not fail")
	}
}
