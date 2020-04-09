package handlerware

import(
	"net/http"
	"golang.org/x/net/context"
)

type BaseHandler    func(http.ResponseWriter, *http.Request)
type ContextHandler func(context.Context, http.ResponseWriter, *http.Request)

var(
	// RequireTls will redirect non-https URLs to the https equiv. Enabled by default.
	RequireTls = true
)

// WithSession is the primary piece of handlerware. It ensures a user is logged in, redirecting
// to a fallback handler if they're not. A usersession is built, and injected into the context.
func WithSession(ch ContextHandler) BaseHandler {
	return WithCtx(EnsureSession(ch))
}

// WithGroup builds on WithSession; it also asserts that the user is a member of a particular group
func WithGroup(g string, ch ContextHandler) BaseHandler {
	return WithCtx(EnsureSession(EnsureGroup(g,ch)))
}

// WithAdmin ensures that the caller has admin privs - either by being
// a user in the AdminGroup, or by having a HTTP header that indicates
// the request came from a trusted part of our appengine world.
func WithAdmin(ch ContextHandler) BaseHandler {
	// EnsureAdmin can succeed even if there is no session, because some
	// X-Appengine headers indicate admin privs. So we set EnsureAdmin
	// as the fallback handler that gets called in the absence of a
	// session - i.e. we call EnsureAdmin either way.
	h := EnsureAdmin(ch)
	return WithCtx(EnsureSessionOrFallback(h, h))
}

/*

import(
	"fmt"
	"net/http"
	"golang.org/x/net/context"
	"github.com/skypies/util/handlerware"
	"github.com/skypies/complaints/config"
)

func init() {
  handlerware.RequireTls = true
  handlerware.TemplateDir = "/app/frontend/web/templates" // Must be relative to module root, i.e. git repo root
  handlerware.InitTemplates()

  handlerware.CtxMakerCallback = func(r *http.Request) context.Context {
    // return context.Context{}
    return r.Context()
  }

  handlerware.CookieName = "serfr0"
  handlerware.InitSessionStore("sekrit", "deadbeef")
  handlerware.NoSessionHandler = func (ctx context.Context, w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/some/relative/URL/for/login", http.StatusFound)
  }


  // 1. Wrapper to ensure user is logged in, injects a UserSession into the context
  http.HandleFunc("/foo",       handlerware.WithSession(fooHandler))


  // 2. Set up the admin group, use the admin wrapper to enforce admin privs
  handlerware.InitGroup(handlerware.AdminGroup, "me@me.com,them@them.com")

  http.HandleFunc("/admin/foo", handlerware.WithAdmin(fooHandler))
  http.HandleFunc("/admin/bar", handlerware.WithAdmin(handlerware.WithoutCtx(baseHandler)))


  // 3. Set up an arbitrary group, enforce membership
  handlerware.InitGroup("my-group", "me@me.com")
  http.HandleFunc("/restricted/foo", handlerware.WithGroup("my-group", fooHandler))
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK\nI am just a normal handler !\n"))
}

func fooHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
  tmpl   := handlerware.GetTemplates(ctx) // will panic() if handler not wired correctly
  sesh,_ := handlerware.GetUserSession(ctx)

	str := fmt.Sprintf("OK\nI have a context !\nI have templates %v!\nI have a user %v!", tmpl, sesh)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(str))
}

 */
