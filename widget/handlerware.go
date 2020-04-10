package widget

/* Common code for pulling out a user session cookie, populating a Context, etc.
import "golang.org/x/net/context"
import "github.com/util/widget"
func init() {
  ctxMakerFunc := func(r *http.Request) context.Context {
    ctx := context.Context{}
  }
  http.HandleFunc("/foo", widget.WithCtx(ctxMakerfunc, fooHandler))
  http.HandleFunc("/bar", widget.WithCtx(widget.WithTemplates(ctxMakerfunc, barHandler)))
  // Note that handlers can be chained, and can set/get values from the context
}
func fooHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf("OK\nI have a context !\n") 
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(str))
}
func barHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
  tmpl := widget.GetTemplates(ctx) // will panic() if handler not wired correctly
	str := fmt.Sprintf("OK\nI have a context !\nI have templates %v!\n", tmpl) 
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(str))
}
 */


import(
	"html/template"
	"net/http"

	"golang.org/x/net/context"
)

type CtxMaker       func(*http.Request) context.Context
type BaseHandler    func(http.ResponseWriter, *http.Request)
type ContextHandler func(context.Context, http.ResponseWriter, *http.Request)

// Outermost wrapper; all other wrappers take (and return) contexthandlers
func WithCtx(f CtxMaker, ch ContextHandler) BaseHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := f(r)
		ch(ctx,w,r)
	}
}

// To prevent other libs colliding with us in the context.Value keyspace, use these private keys
type contextKey int
const(
	templatesKey contextKey = iota
)

// WithFoo: injects relevant object into the context
func WithTemplates(t *template.Template, ch ContextHandler) ContextHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, templatesKey, t)		
		ch(ctx, w, r)
	}
}
// GetFoo: given a context, extracts the object (or panics; should not be optional)
func GetTemplates(ctx context.Context) (*template.Template) {
	tmpl, ok := ctx.Value(templatesKey).(*template.Template)
	if (!ok) { panic ("handlerware.GetTemplates: no object found in context") }
	return tmpl
}

func WithCtxTmpl(f CtxMaker, t *template.Template, ch ContextHandler) BaseHandler {
	return WithCtx(f, WithTemplates(t, ch))
}
