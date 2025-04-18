package ds

import(
	"errors"
	"net/http"

	"context"
)

var(
	ErrNoSuchEntity = errors.New("dsprovider: no such entity")
	ErrFieldMismatch = errors.New("dsprovider: src obj had a field that dst obj didn't")
	ErrNoMemcacheService = errors.New("dsprovider: no memcache service available")
)

// Keyer is a very thin wrapper. It should be populated with a *datastore.Key
type Keyer interface {
	Encode() string
}

// Provider is a wrapper over the datastore APIs (cloud and appengine), so that
// client code (and in particular query assembley) can run both inside and outside of google
// appengine.
//
// In fact, it wraps all the appengine APIs (incl. /log and /urlfetch), so we don't
// need to worry about any appengine libs except in util/ae/ds.
type DatastoreProvider interface {
	Get(ctx context.Context, keyer Keyer, dst interface{}) error
	GetMulti(ctx context.Context, keyers []Keyer, dst interface{}) error
	GetAll(ctx context.Context, q *Query, dst interface{}) ([]Keyer, error)
	Put(ctx context.Context, keyer Keyer, src interface{}) (Keyer, error)
	PutMulti(ctx context.Context, keyers []Keyer, src interface{}) ([]Keyer, error)
	Delete(ctx context.Context, keyer Keyer) error
	DeleteMulti(ctx context.Context, keyers []Keyer) error
	
	NewIncompleteKey(ctx context.Context, kind string, root Keyer) Keyer
	NewNameKey(ctx context.Context, kind, name string, root Keyer) Keyer
	NewIDKey(ctx context.Context, kind string, id int64, root Keyer) Keyer
	DecodeKey(encoded string) (Keyer, error)
	KeyParent(Keyer) Keyer
	KeyName(Keyer) string

	// HTTP client - maybe urlfetch, maybe not
	HTTPClient(ctx context.Context) *http.Client

	// Logging support - goes to appengine logging, or maybe STDOUT
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warningf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Criticalf(ctx context.Context, format string, args ...interface{})
}
