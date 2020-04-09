package handlerware

// Routines that work with the context, and do all our setup (object injection etc)

import(
	"net/http"

	"golang.org/x/net/context"
)

type CtxMaker       func(*http.Request) context.Context

var(
	// CtxMakerCallback creates a context.Context as the app would like (i.e. with timeouts etc.)
	// The user can override this via their init() block, but the default isn't terrible.
	CtxMakerCallback = func(r *http.Request) context.Context {
		return r.Context()
	}
)

// To prevent other libs colliding with us in the context.Value keyspace, use this private type and keys
type contextKey int
const(	
	sessionKey contextKey = iota
	templatesKey
)

// WithCtx is the outermost wrapper, which returns a BaseHandler
// suitable for http.HandleFunc; the rest of the handlerware works on
// ContextHandlers, and can be chained. This handler will enforce TLS
// if needed, create the context.Context, and inject the templates
// into that context, before calling whatever is next in the chain.
func WithCtx(ch ContextHandler) BaseHandler {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check for https, if we need to
		if RequireTls && r.Header.Get("x-appengine-https") == "off" {
			new := r.URL
			new.Scheme = "https"
			new.Host = r.Host // r.URL is weirdly unpopulated, so copy over the hostname
			http.Redirect(w, r, new.String(), http.StatusFound)
			return
		}
		
		ctx := CtxMakerCallback(r)

		// Inject the templates (may be nil, whatever)
		ctx = context.WithValue(ctx, templatesKey, Templates)

		ch(ctx,w,r)
	}
}

// WithoutCtx lets us strip out the context, so we can wrap regular BaseHandlers. This is
// mostly just so internal cron URLs can be wrapped inside WithAdmin().
func WithoutCtx(bh BaseHandler) ContextHandler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		bh(w,r)
	}
}
