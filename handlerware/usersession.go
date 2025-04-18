package handlerware

import(
	"fmt"
	"log"
	"net/http"
	//"net/http/httputil"
	"time"

	"context"

	gsessions "github.com/gorilla/sessions"
)

// FIXME: logging needs a real fix
func logPrintf(r *http.Request, fmtstr string, varargs ...interface{}) {
	payload := fmt.Sprintf(fmtstr, varargs...)
	prefix := fmt.Sprintf("ip:%s", r.Header.Get("x-appengine-user-ip"))
	log.Printf("%s %s", prefix, payload)
}

var(
	// CookieName is what the calling app wants its session token to be kept in.
	CookieName = "choc_chip"

	// NoSessionHandler is executed when the user doesn't have a session.
	NoSessionHandler ContextHandler
	
	sessionStore *gsessions.CookieStore
)

// InitSessionStore *must* be called in the caller's init() block.
func InitSessionStore(key, prevkey string) {
	sessionStore = gsessions.NewCookieStore(
		[]byte(key), nil,
		[]byte(prevkey), nil)

	sessionStore.MaxAge(86400 * 180)
}

// Pretty much all handlers should expect to be able to pluck this object out of their
// Context; see handlerware.go
type UserSession struct {
	Email        string          // case sensitive, sadly
	CreatedAt    time.Time       // when the user last went through the OAuth2 dance
}

func (us UserSession)IsEmpty() bool { return us.Email == "" }
func (us UserSession)IsAdmin() bool { return us.IsInGroup(AdminGroup) }
func (us UserSession)IsInGroup(g string) bool { return !us.IsEmpty() && IsInGroup(g,us.Email) }

// {{{ EnsureSession{OrFallback}

// EnsureSession checks that there is a user session, and if so runs the
// specified handler; else it runs the `NoSessionHandler` (which presumably
// starts a login flow). Adds some debug logging into a cookie, to try and
// illuminate how users end up without sessions.
func EnsureSession(ch ContextHandler) ContextHandler {
	return EnsureSessionOrFallback(ch, NoSessionHandler)
}

// EnsureSessionOrFallback lets the caller specify which contexthandler
// to run when the session is not found.
func EnsureSessionOrFallback(ch,fallback ContextHandler) ContextHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		crumbs := CrumbTrail{}
		crumbCookieName := CookieName + "crumbs"

		// First, extract prev breadcrumbs and log them
		cookies := map[string]string{}
		for _,c := range r.Cookies() {
			crumbs.Add("C:"+c.Name)
			cookies[c.Name] = c.Value
		}
		//if val,exists := cookies[crumbCookieName]; exists {
		//	logPrintf(r, "%s in : %s", crumbCookieName, val)
		//}

		handler := fallback
		user := "??"

		if _,exists := cookies[CookieName]; exists {
			sesh,err := req2Session(r, &crumbs)
			if err == nil && !sesh.IsEmpty() {
				user = sesh.Email
				// Stash the session in the context, and move on to the proper handler
				ctx = setUserSession(ctx, sesh)
				handler = ch

			} else {
				if err != nil { logPrintf(r, "req2session err: " + err.Error()) }
				logPrintf(r, "crumbs: " + crumbs.String())
			}

		} else {
			crumbs.Add("NoMainCookie")
		}
		
		// Before invoking final handler, log breadcrumb trail, and stash in cookie
		//logPrintf(r, "%s out: %s", crumbCookieName, crumbs)
		cookie := http.Cookie{
			Name: crumbCookieName,
			Value: crumbs.String(),
			Expires:time.Now().AddDate(1,0,0),
		}
		http.SetCookie(w, &cookie)

		//reqLog,_ := httputil.DumpRequest(r,true)
		//logPrintf(r, "HTTP req>>>>\n%s====\n", reqLog)
		
		if handler == nil {
			logPrintf(r, "WithSession had no session, no NoSessionHandler")
			http.Error(w, fmt.Sprintf("no session, no NoSessionHandler (%s)", r.URL), http.StatusInternalServerError)
			return
		}

		logPrintf(r, "userSession(%s): %s", user, r.URL)

		handler(ctx, w, r)
	}
}

// }}}
// {{{ {Get,set}UserSession

func GetUserSession(ctx context.Context) (UserSession, bool) {
	opt, ok := ctx.Value(sessionKey).(UserSession)
	return opt, ok
}

func setUserSession(ctx context.Context, sesh UserSession) context.Context {
	return context.WithValue(ctx, sessionKey, sesh)
}

// }}}

// {{{ CreateSession

func CreateSession(ctx context.Context, w http.ResponseWriter, r *http.Request, sesh UserSession) {
	session,err := sessionStore.Get(r, CookieName)
	if err != nil {
		// This isn't usually an important error (the session was most likely expired, which is why
		// we're logging in) - so log as Info, not Error.
		log.Printf("CreateSession: sessionStore.Get [failing is OK for this call] had err: %v", err)
	}

	session.Values["email"] = sesh.Email
	session.Values["tstamp"] = time.Now().Format(time.RFC3339)
	if err := session.Save(r,w); err != nil {
		log.Printf("CreateSession: session.Save: %v", err)
	}
	log.Printf("CreateSession OK for %s", sesh.Email)
}

// }}}\
// {{{ OverwriteSessionToNil

func OverwriteSessionToNil(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Ignore errors; we just want an empty one
	session,_ := sessionStore.Get(r, CookieName)

	session.Values["email"] = nil
	session.Values["tstamp"] = nil

	session.Save(r, w)	
	log.Printf("OverwriteSessionToNil done")
}

// }}}

// {{{ req2Session

func req2Session(r *http.Request, crumbs *CrumbTrail) (UserSession, error) {
	// If not found, returns an empty session
	session,err := sessionStore.Get(r, CookieName)
	if err != nil {
		crumbs.Add("GDecodeFailed")
		return UserSession{}, fmt.Errorf("Req2Session: sessionStore.Get: %v", err)
	}

	if session.IsNew {
		crumbs.Add("NewGSession")
		return UserSession{}, nil

	} else if session.Values["email"] == nil {
		crumbs.Add("LoggedOutSession")
		return UserSession{}, nil
	}

	crumbs.Add("SessionRetrieved")

	// crumbs.Add("E:"+session.Values["email"].(string))
	
	tstampStr := session.Values["tstamp"].(string)
	tstamp,_ := time.Parse(time.RFC3339, tstampStr)

	crumbs.Add(fmt.Sprintf("Age:%s", time.Since(tstamp)))
	
	userSesh := UserSession{
		Email: session.Values["email"].(string),
		CreatedAt: tstamp,
	}

	// In case of a new session object, give it a long cookie lifetime
	//session.Options.MaxAge = 86400 * 180 // Default is 4w.

	return userSesh, nil
}

// }}}

// {{{ -------------------------={ E N D }=----------------------------------

// Local variables:
// folded-file: t
// end:

// }}}
