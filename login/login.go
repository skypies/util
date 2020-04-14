package login

// Generic oauth2 stuff, and impls for Google and Facebook

/* How to use it.

1. Decide on your URLs.

Say you have https://site.com/.
You can 'mount' the login URLs onto it, say at: site.com/foo/login
The bare /foo/login
Each supported OAuth backend will register a URL under there (e.g. site.com/foo/login/google)

2. Setup your Google Cloud Project (or Facebook dev account) appropriately.
   For GCP, this is under 'APIs & Services'.
2a. It'll need an "oauth consent" screen to be setup
2b. It'll need an "OAuth 2.0 Client ID", of type "Web application", under credentials
2c. The 'javascript origins' should be https://site.com, and perhaps https://thing.appspot.com, etc
2d. The 'redirect URLs' should be site.com/foo/login/google (and perhaps for localhsot too)

3. Get the AppId and Secret for the credential (2b above), and put it somewhere in the config for your app.

4. Somewhere in app/frontend/main.go:init(), setup the globals, and init:
	login.Host                  = "https://site.com"
	login.RedirectUrlStem       = "/foo/login" // oauth2 callbacks will register  under here
	login.AfterLoginRelativeUrl = "/foo/home" // where the user finally ends up, after being logged in
	login.GoogleClientID        = config.Get("google.oauth2.appid")
	login.GoogleClientSecret    = config.Get("google.oauth2.secret")
  ...
	login.Init()

5. create a loginPageHandler (and its template) a bit like this:

	func loginPageHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {	
		templates := hw.GetTemplates(ctx)
		var params = map[string]interface{}{
			"google": login.Goauth2.GetLoginUrl(w,r),
			"googlefromscratch": login.Goauth2.GetLogoutUrl(w,r),
		}
		if err := templates.ExecuteTemplate(w, "login", params); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	{{define "login"}}
	<html>
		{{template "header"}}
		<body>
			<div class="stack">
				<h1>The FDB system needs a Google login</h1>      
				<div class="stack">
					<div><a href="{{.google}}"><img width="300" src="/static/google-signin.png"/></a></div>
				</div>
				<p>(PS: if you're a Google power-user with multiple accounts,
				<a href="{{.googlefromscratch}}">click here</a> to logout and
				then log back into the account you want to use on stop.jetnoise.)</p>
			</div>
		</body>
	</html>
	{{end}}

7. Configure the util/handlerware

  hw.NoSessionHandler = loginPageHandler // i.e. where to start the login flow
  hw.CookieName = "somecookie"
  hw.InitSessionStore(config.Get("sessions.key"), config.Get("sessions.prevkey")) // encryption for the cookie

  // This should really get done behind the scenes, but hey
	login.OnSuccessCallback = func(w http.ResponseWriter, r *http.Request, email string) error {
		hw.CreateSession(r.Context(), w, r, hw.UserSession{Email:email})
		return nil
	}

8. Protect some URLs, using hw.WithSession() or hw.WithAdmin()



 */

import(
	"log"
	"net/http"
)

type Oauth2er interface {
	Name() string
	GetLoginUrl(w http.ResponseWriter, r *http.Request) string
	GetLogoutUrl(w http.ResponseWriter, r *http.Request) string
	CallbackToEmail(r *http.Request) (string, error)
}

func getCallbackRelativeUrl(o Oauth2er) string {
	return RedirectUrlStem + "/" +	o.Name()
}

type Oauth2SucessCallback func(w http.ResponseWriter, r *http.Request, email string) error

var(
	// The caller should configure these values. Note that the URLs need
	// to be added to your Google and Facebook setups, as allowed
	// referrers.
	Host                        = "https://stop.jetnoise.net"
	RedirectUrlStem             = "/login" // oauth2 callbacks will register  under here
	AfterLoginRelativeUrl       = "/" // where the user finally ends up, after being logged in
	OnSuccessCallback           Oauth2SucessCallback
	
	// Individual oauth2 systems
	Goauth2 Oauth2er
	Fboauth2 Oauth2er
)

// The caller *must* call this, after they've set the vars above
func Init() {
	Goauth2  = NewGoogleOauth2()
	Fboauth2 = NewFacebookOauth2()

	http.HandleFunc(getCallbackRelativeUrl(Goauth2), NewOauth2Handler(Goauth2))
	http.HandleFunc(getCallbackRelativeUrl(Fboauth2), NewOauth2Handler(Fboauth2))
}

// Returns a standard requesthandler. When run, it handles the redirect from the oauth2
// provider, and if it gets an email address from the provider, will invoke the
// callback function with it, before redirecting to the provided URL.
func NewOauth2Handler(oauth2 Oauth2er) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if email,err := oauth2.CallbackToEmail(r); err != nil {
			log.Printf("NewOauth2Handler/CallbackToEmail: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		} else {
			if OnSuccessCallback != nil {
				if err := OnSuccessCallback(w,r,email); err != nil {
					log.Printf("NewOauth2Handler/OnSuccessCallback: %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// http.Redirect(w, r, AfterLoginRelativeUrl +"#email:"+email, http.StatusFound)
			http.Redirect(w, r, AfterLoginRelativeUrl, http.StatusFound)
		}
	}
}

// Helper, because r.URL is weirdly unpopulated most of the time
func makeUrlAbsolute(r *http.Request, relativePath string) string {
	new := r.URL
	new.Path = relativePath

	if new.Scheme == "" { new.Scheme = "https" }
	if new.Host == "" { new.Host = r.Host }

	return new.String()
}
