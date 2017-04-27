package resources

import (
	"errors"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// contains specific constant names for usage in pkg.
const (
	ResponsePerPageName = "total"
	PerPageName         = "page"
)

// Guarded defines a struct which exposes a session secured request life cycle where a request made will be guarded
// with specific data from a underline session and will be validated when receiving response.
type Guarded struct {
	SessionName  string
	CookieName   string
	CookieSecret string
	Cookies      sessions.CookieStore
	Log          sink.Sink
}

// Guard attempts to added incoming request with a session which is stored in the outgoing response which
// then will be used to guard against other incoming request.
func (u Guarded) Guard(w http.ResponseWriter, r *http.Request) error {
	defer u.Log.Emit(sinks.Info("Guard Request").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"path":   r.URL.Path,
	}).Trace("Guarded.Guard").End())

	defer context.Clear(r)

	session, err := u.Cookies.Get(r, u.SessionName)
	if err != nil {
		u.Log.Emit(sinks.Error("Cookie Retreival Failed: %+q", err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}))

		return err
	}

	session.Values[u.CookieName] = u.CookieSecret
	session.Save(r, w)

	return nil
}

// Validate attempts to authenticate incoming request with a sessio data expected from the request.
func (u Guarded) Validate(w http.ResponseWriter, r *http.Request) error {
	defer u.Log.Emit(sinks.Info("Validated Guarded Request").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"path":   r.URL.Path,
	}).Trace("Guarded.Validate").End())

	defer context.Clear(r)

	session, err := u.Cookies.Get(r, u.SessionName)
	if err != nil {
		u.Log.Emit(sinks.Error("Cookie Retreival Failed: %+q", err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}))

		return err
	}

	// Attempt to retrieve specific Guard.CookieName in retrieved session.
	value, ok := session.Values[u.CookieName]
	if !ok {
		err := errors.New("Session cookie guard not found in request")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}))

		return err
	}

	// Did value match expected guard secret?
	if value != u.CookieSecret {
		err := errors.New("Session cookie guard not found in request")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}))

		return err
	}

	return nil
}
