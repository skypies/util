package memcache

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	//"github.com/skypies/util/singleton"
	mc "github.com/skypies/util/singleton/memcache"
)

type Foo struct {
	S string
}

var (
	ctx = context.Background()
	server = "192.168.80.33:11211"
)

func test_small() {
	str := "Splendid. "
	
	name := "foo_3"
	foo1 := Foo{S:str}
	foo2 := Foo{}

	fmt.Printf("== Small tests ==\n")
	p := mc.NewProvider(server)

	if err := p.WriteSingleton(ctx, name, &foo1); err != nil {
		fmt.Printf(" * err Write %s: %v\n", name, err)
	}
	fmt.Printf("Wrote foo1: %d bytes\n", len(foo1.S))
	
	if err := p.ReadSingleton(ctx, name, &foo2); err != nil {
		fmt.Printf(" * err Read %s: %v\n", name, err)
	}
	fmt.Printf("Read foo2 : %d bytes\n", len(foo2.S))
}


func test_big() {
	str := "Splendid. "
	bigStr := strings.Repeat(str, 1300000)
	
	name := "foo_3"
	foo1 := Foo{S:bigStr}
	foo2 := Foo{}

	p := mc.NewProvider(server)
	p.Timeout = time.Second * 4
	p.ShardCount = 32

	fmt.Printf("== Big tests ==\n")

	if err := p.WriteSingleton(ctx, name, &foo1); err != nil {
		fmt.Printf(" * err Write %s: %v\n", name, err)
	}
	fmt.Printf("Wrote foo1: %d bytes\n", len(foo1.S))
	
	if err := p.ReadSingleton(ctx, name, &foo2); err != nil {
		fmt.Printf(" * err Read %s: %v\n", name, err)
	}
	fmt.Printf("Read foo2 : %d bytes\n", len(foo2.S))
}

func main() {	
	fmt.Printf("Hello !\n")

	test_small()
	test_big()
}
