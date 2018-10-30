package ds

import(
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"cloud.google.com/go/datastore"
)

var Debug = false

// CloudDSProvider implements the DatastoreProvider interface using the cloud datastore API,
// for use outside of appengine environments.
type CloudDSProvider struct {
	Project    string
	client    *datastore.Client
}

func NewCloudDSProvider(ctx context.Context, project string) (*CloudDSProvider, error) {
	client,err := datastore.NewClient(ctx, project,
		option.WithGRPCDialOption(grpc.WithBackoffMaxDelay(5*time.Second)),
		option.WithGRPCDialOption(grpc.WithBlock()),
		option.WithGRPCDialOption(grpc.WithTimeout(30*time.Second)))
	provider := CloudDSProvider{Project: project, client: client}	

	log.SetOutput(os.Stdout)

	return &provider, err
}

func (p CloudDSProvider)flattenQuery(in *Query) *datastore.Query {
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

func  (p CloudDSProvider)unpackKeyer(in Keyer) *datastore.Key {
	if in == nil { return nil }
	return in.(*datastore.Key)
}
func (p CloudDSProvider)unpackKeyers(in []Keyer) []*datastore.Key {
	out := []*datastore.Key{}
	for _,keyer := range in {
		out = append(out, p.unpackKeyer(keyer))
	}
	return out
}
func (p CloudDSProvider)packKeyers(in []*datastore.Key) []Keyer {
	out := []Keyer{}
	for _,k := range in {
		out = append(out, Keyer(k))
	}
	return out
}

func (p CloudDSProvider)GetAll(ctx context.Context, q *Query, dst interface{}) ([]Keyer, error) {
	dsQuery := p.flattenQuery(q)

	keys,err := p.client.GetAll(ctx, dsQuery, dst)
	keyers := p.packKeyers(keys)

	if err != nil {
		if _,assertionOk := err.(*datastore.ErrFieldMismatch); assertionOk {
			return keyers, ErrFieldMismatch
		}
		return nil, fmt.Errorf("GetAll{cloud}: %v\nQuery: %s", err, q)
	}
	return keyers,nil
}

func (p CloudDSProvider)Get(ctx context.Context, keyer Keyer, dst interface{}) error {
	err := p.client.Get(ctx, p.unpackKeyer(keyer), dst)
	if err == datastore.ErrNoSuchEntity {
		return ErrNoSuchEntity
	} else if err != nil {
		if _,assertionOk := err.(*datastore.ErrFieldMismatch); assertionOk {
			return ErrFieldMismatch
		}
	}
	return err
}

func (p CloudDSProvider)GetMulti(ctx context.Context, keyers []Keyer, dst interface{}) error {
	err := p.client.GetMulti(ctx, p.unpackKeyers(keyers), dst)
	if err == datastore.ErrNoSuchEntity {
		return ErrNoSuchEntity
	} else if err != nil {
		if _,assertionOk := err.(*datastore.ErrFieldMismatch); assertionOk {
			return ErrFieldMismatch
		}
	}
	return err
}

func (p CloudDSProvider)Put(ctx context.Context, keyer Keyer, src interface{}) (Keyer, error) {
	key,error := p.client.Put(ctx, p.unpackKeyer(keyer), src)
	return Keyer(key), error
}	
func (p CloudDSProvider)PutMulti(ctx context.Context, keyers []Keyer, src interface{}) ([]Keyer, error) {
	keys,err := p.client.PutMulti(ctx, p.unpackKeyers(keyers), src)
	return p.packKeyers(keys), err
}
func (p CloudDSProvider)Delete(ctx context.Context, keyer Keyer) error {
	err := p.client.Delete(ctx, p.unpackKeyer(keyer))
	if err ==	datastore.ErrNoSuchEntity { return ErrNoSuchEntity }
	return err
}	
func (p CloudDSProvider)DeleteMulti(ctx context.Context, keyers []Keyer) error {
	err := p.client.DeleteMulti(ctx, p.unpackKeyers(keyers))
	if err ==	datastore.ErrNoSuchEntity { return ErrNoSuchEntity }
	return err
}

func (p CloudDSProvider)NewIncompleteKey(ctx context.Context, kind string, root Keyer) Keyer {
	key := datastore.IncompleteKey(kind, p.unpackKeyer(root))
	return Keyer(key)
}
func (p CloudDSProvider)NewNameKey(ctx context.Context, kind, name string, root Keyer) Keyer {
	key := datastore.NameKey(kind, name, p.unpackKeyer(root))
	return Keyer(key)
}
func (p CloudDSProvider)NewIDKey(ctx context.Context, kind string, id int64, root Keyer) Keyer {
	key := datastore.IDKey(kind, id, p.unpackKeyer(root))
	return Keyer(key)
}

func (p CloudDSProvider)DecodeKey(encoded string) (Keyer, error) {
	key, err := datastore.DecodeKey(encoded)
	return Keyer(key), err
}
func (p CloudDSProvider)KeyParent(in Keyer) Keyer {
	if parentKey := p.unpackKeyer(in).Parent; parentKey != nil {
		return Keyer(parentKey)
	}
	return nil
}
func (p CloudDSProvider)KeyName(in Keyer) string { return p.unpackKeyer(in).Name }


func (p CloudDSProvider)HTTPClient(ctx context.Context) *http.Client {
	c := http.Client{}
	return &c
}



func (p CloudDSProvider)Debugf(ctx context.Context, format string, args ...interface{}) {
	if Debug {log.Printf(format, args...)}
}
func (p CloudDSProvider)Infof(ctx context.Context, format string,args ...interface{}) {
	log.Printf(format, args...)
}
func (p CloudDSProvider)Errorf(ctx context.Context, format string,args ...interface{}) {
	log.Printf(format, args...)
}
func (p CloudDSProvider)Warningf(ctx context.Context, format string,args ...interface{}) {
	log.Printf(format, args...)
}
func (p CloudDSProvider)Criticalf(ctx context.Context, format string,args ...interface{}) {
	log.Printf(format, args...)
}


/*
func (p CloudDSProvider)MemcacheGet(ctx context.Context, name string) ([]byte, error) {
	return nil, ErrNoMemcacheService
}
func (p CloudDSProvider)MemcacheSet(ctx context.Context, name string, data []byte) error {
	return ErrNoMemcacheService
}	
func (p CloudDSProvider)MemcacheDelete(ctx context.Context, name string) error {
	return ErrNoMemcacheService
}
*/
