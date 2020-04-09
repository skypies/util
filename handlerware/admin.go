package handlerware

import(
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

// EnsureGroup asserts that the user is logged in, and is a member of
// the specified group; if not, then 401.
func EnsureGroup(g string, ch ContextHandler) ContextHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		sesh,hadSesh := GetUserSession(ctx)

		if hadSesh && IsInGroup(g, sesh.Email) {
			// We're in the group - run the handler
			ch(ctx, w, r)

		} else {
			errstr :=  "This URL requires you to be logged in"
			if hadSesh {
				errstr = fmt.Sprintf("This URL requires you to be in the group %q", g)
			}
			http.Error(w, errstr, http.StatusUnauthorized)
		}
	}
}

// EnsureAdmin validates that the request has admin privileges, and runs the handler (or returns 401).
// Privileges are either that the user is logged in, and is an admin; or that the request came from
// an appengine cron job or a cloud tasks queue.
// (see https://cloud.google.com/appengine/docs/flexible/nodejs/scheduling-jobs-with-cron-yaml#validating_cron_requests,
// https://cloud.google.com/tasks/docs/creating-appengine-handlers#reading_app_engine_task_request_headers)
func EnsureAdmin(ch ContextHandler) ContextHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fromCron := r.Header.Get("x-appengine-cron")
		fromCloudTasks := r.Header.Get("x-appengine-queuename")

		sesh,hadSesh := GetUserSession(ctx)

		haveAdmin := false
		switch {
		case fromCron != "":                                haveAdmin = true
		case fromCloudTasks != "":                          haveAdmin = true
		case hadSesh && IsInGroup(AdminGroup, sesh.Email):  haveAdmin = true
		default:                                            haveAdmin = false
		}

		if !haveAdmin {
			errstr := "This URL requires you to be logged in"
			if hadSesh {
				errstr = "This URL requires admin access"
			}
			http.Error(w, errstr, http.StatusUnauthorized)
			return
		}
				
		// We have admin rights - run the handler
		ch(ctx,w,r)
	}
}

