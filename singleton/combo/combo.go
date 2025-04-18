package combo

// Combines two singleton providers - a primary and a secondary.

/*

 // The logic is to retrieve from primary first, and secondary if primaruy fails; and to
 // always write to secondary, but ignore write fails to primary.
 // This is designed for {memcache,datastore} tiering - we transparently benefit from
 // memcache, but have guaranteed persistence into datastore.

 import(
   "github.com/skypies/util/singleton"
   "github.com/skypies/util/singleton/memcache"
   dssingleton "github.com/skypies/util/gcp/singleton"
 )

 p1 := memcache.NewProvider(...)
 p2 := dssingleton.NewProvider(...)

 p := combo.NewProvider(p1, p2)

*/

import(
	"context"
	"github.com/skypies/util/singleton"
)

type ComboSingletonProvider struct {
	Primary singleton.SingletonProvider
	Secondary singleton.SingletonProvider
}

func NewProvider(primary,secondary singleton.SingletonProvider) ComboSingletonProvider {
	return ComboSingletonProvider{
		Primary: primary,
		Secondary: secondary,
	}
}

func (sp ComboSingletonProvider)ReadSingleton(ctx context.Context, name string, f singleton.NewReaderFunc, ptr interface{}) error {
	if err := sp.Primary.ReadSingleton(ctx, name, f, ptr); err != nil {
		// Ignore primary error, and fall back to secondary
		return sp.Secondary.ReadSingleton(ctx, name, f, ptr)
	}

	return nil
}

func (sp ComboSingletonProvider)WriteSingleton(ctx context.Context, name string, f singleton.NewWriteCloserFunc, ptr interface{}) error {
	if err := sp.Secondary.WriteSingleton(ctx, name, f, ptr); err != nil {
		// Secondary write errors are fatal
		return err
	}

	// Ignore primary's error
	sp.Primary.WriteSingleton(ctx, name, f, ptr)

	return nil 
}
