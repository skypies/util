package ds

// https://cloud.google.com/appengine/docs/standard/go/datastore/reference

import(
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"

	base "github.com/skypies/util/gcp/ds"
)

var Debug = false

// AppengineDSProvider implements the DatastoreProvider interface using the appengine datastore API,
// for use inside appengine environments.
type AppengineDSProvider struct {
}

func (p AppengineDSProvider)FlattenQuery(in *base.Query) *datastore.Query {
	out := datastore.NewQuery(in.Kind)
	if in.AncestorKeyer != nil { out = out.Ancestor(in.AncestorKeyer.(*datastore.Key)) }
	for _,filter := range in.Filters {
		out = out.Filter(filter.Field, filter.Value)
	}
	if len(in.ProjectFields) != 0 { out = out.Project(in.ProjectFields...) }
	if in.OrderStr != ""          { out = out.Order(in.OrderStr) }
	if in.KeysOnlyVal             { out = out.KeysOnly() }
	if in.DistinctVals            { out = out.Distinct() }
	if in.LimitVal != 0           { out = out.Limit(in.LimitVal) }
	return out
}

func (p AppengineDSProvider)unpackKeyer(in base.Keyer) *datastore.Key {
	if in == nil { return nil }
	return in.(*datastore.Key)
}
func (p AppengineDSProvider)unpackKeyers(in []base.Keyer) []*datastore.Key {
	out := []*datastore.Key{}
	for _,keyer := range in {
		out = append(out, p.unpackKeyer(keyer))
	}
	return out
}
func (p AppengineDSProvider)packKeyers(in []*datastore.Key) []base.Keyer {
	out := []base.Keyer{}
	for _,k := range in {
		out = append(out, base.Keyer(k))
	}
	return out
}


func (p AppengineDSProvider)GetAll(ctx context.Context, q *base.Query, dst interface{}) ([]base.Keyer, error) {
	aeQuery := p.FlattenQuery(q)	
	keys,err := aeQuery.GetAll(ctx, dst)
	keyers := p.packKeyers(keys)

	if err != nil {
		if _,assertionOk := err.(*datastore.ErrFieldMismatch); assertionOk {
			return keyers, base.ErrFieldMismatch
		}
		return nil, fmt.Errorf("GetAll{AE}: %v\nQuery: %s", err, q)
	}

	return keyers,nil
}

func (p AppengineDSProvider)Get(ctx context.Context, keyer base.Keyer, dst interface{}) error {
	err := datastore.Get(ctx, p.unpackKeyer(keyer), dst)

	if err == datastore.ErrNoSuchEntity {
		return base.ErrNoSuchEntity
	} else if err != nil {
		if _,assertionOk := err.(*datastore.ErrFieldMismatch); assertionOk {
			return base.ErrFieldMismatch
		}
	}
	return err
}

func (p AppengineDSProvider)GetMulti(ctx context.Context, keyers []base.Keyer, dst interface{}) error {
	err := datastore.GetMulti(ctx, p.unpackKeyers(keyers), dst)

	if err == datastore.ErrNoSuchEntity {
		return base.ErrNoSuchEntity
	} else if err != nil {
		if _,assertionOk := err.(*datastore.ErrFieldMismatch); assertionOk {
			return base.ErrFieldMismatch
		}
	}
	return err
}

func (p AppengineDSProvider)Put(ctx context.Context, keyer base.Keyer, src interface{}) (base.Keyer, error) {
	key,error := datastore.Put(ctx, p.unpackKeyer(keyer), src)
	return base.Keyer(key), error
}	
func (p AppengineDSProvider)PutMulti(ctx context.Context, keyers []base.Keyer, src interface{}) ([]base.Keyer, error) {
	keys,err := datastore.PutMulti(ctx, p.unpackKeyers(keyers), src)
	return p.packKeyers(keys), err
}
func (p AppengineDSProvider)Delete(ctx context.Context, keyer base.Keyer) error {
	err := datastore.Delete(ctx, p.unpackKeyer(keyer))
	if err ==	datastore.ErrNoSuchEntity { return base.ErrNoSuchEntity }
	return err
}	
func (p AppengineDSProvider)DeleteMulti(ctx context.Context, keyers []base.Keyer) error {
	err := datastore.DeleteMulti(ctx, p.unpackKeyers(keyers))
	if err ==	datastore.ErrNoSuchEntity { return base.ErrNoSuchEntity }
	return err
}	

func (p AppengineDSProvider)NewIncompleteKey(ctx context.Context, kind string, root base.Keyer) base.Keyer {
	key := datastore.NewIncompleteKey(ctx, kind, p.unpackKeyer(root))
	return base.Keyer(key)
}
func (p AppengineDSProvider)NewNameKey(ctx context.Context, kind, name string, root base.Keyer) base.Keyer {
	key := datastore.NewKey(ctx, kind, name, 0, p.unpackKeyer(root))
	return base.Keyer(key)
}
func (p AppengineDSProvider)NewIDKey(ctx context.Context, kind string, id int64, root base.Keyer) base.Keyer {
	key := datastore.NewKey(ctx, kind, "", id, p.unpackKeyer(root))
	return base.Keyer(key)
}

func (p AppengineDSProvider)DecodeKey(encoded string) (base.Keyer, error) {
	key, err := datastore.DecodeKey(encoded)
	return base.Keyer(key), err
}
func (p AppengineDSProvider)KeyParent(in base.Keyer) base.Keyer {
	if parentKey := p.unpackKeyer(in).Parent(); parentKey != nil {
		return base.Keyer(parentKey)
	}
	return nil
}
func (p AppengineDSProvider)KeyName(in base.Keyer) string { return p.unpackKeyer(in).StringID() }


func (p AppengineDSProvider)HTTPClient(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}


func (p AppengineDSProvider)Debugf(ctx context.Context, format string, args ...interface{}) {
	if Debug {log.Debugf(ctx, format, args...)}
}
func (p AppengineDSProvider)Infof(ctx context.Context, format string,args ...interface{}) {
	log.Infof(ctx, format, args...)
}
func (p AppengineDSProvider)Errorf(ctx context.Context, format string,args ...interface{}) {
	log.Errorf(ctx, format, args...)
}
func (p AppengineDSProvider)Warningf(ctx context.Context, format string,args ...interface{}) {
	log.Warningf(ctx, format, args...)
}
func (p AppengineDSProvider)Criticalf(ctx context.Context, format string,args ...interface{}) {
	log.Criticalf(ctx, format, args...)
}

// This function can be used in handlerware to create a context
func CtxMakerFunc(r *http.Request) context.Context {
	ctx,_ := context.WithTimeout(appengine.NewContext(r), 550 * time.Second)
	return ctx
}


/*
func (p AppengineDSProvider)MemcacheGet(ctx context.Context, name string) ([]byte, error) {
	item,err := memcache.Get(ctx, name)
	if err != nil {
    return nil, err  // might be memcache.ErrCacheMiss
	}
	return item.Value, nil
}

func (p AppengineDSProvider)MemcacheSet(ctx context.Context, name string, data []byte) error {
	if len(data) > 950000 {
		return fmt.Errorf("memecache object too large (name=%s, size=%d)", name, len(data))
	}
	item := &memcache.Item{Key:name, Value:data}
	return memcache.Set(ctx, item)
}	

func (p AppengineDSProvider)MemcacheDelete(ctx context.Context, name string) error {
	return memcache.Delete(ctx, name)
}
*/
