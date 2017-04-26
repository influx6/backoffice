package resources

import (
	"errors"
	"net/http"

	"github.com/influx6/backoffice/auth"
	"github.com/influx6/backoffice/handlers"
	"github.com/influx6/backoffice/utils"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
	"golang.org/x/oauth2"
)

// Auth defines an handler which provides authorization handling for
// a request, needing user authentication.
type Auth struct {
	handlers.BearerAuth
}

// CheckAuthorization handles receiving requests to verify user authorization.
/* Service API
HTTP Method: GET
Header:
		{
			"Authorization":"Bearer <TOKEN>",
		}

		WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>
*/
func (u Auth) CheckAuthorization(w http.ResponseWriter, r *http.Request, params map[string]string) error {
	defer u.Log.Emit(sinks.Info("Authenticate Authorization").WithFields(sink.Fields{
		"params": params,
		"remote": r.RemoteAddr,
		"path":   r.URL.Path,
	}).Trace("Auth.CheckAuthorization").End())

	// Retrieve authorization header.
	if err := u.BearerAuth.CheckAuthorization(r.Header.Get("Authorization")); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		http.Error(w, utils.ErrorMessage(http.StatusInternalServerError, "Invalid Auth: Failed to validate authorization", err), http.StatusInternalServerError)
		return err
	}

	return nil
}

//==================================================================================================================================================================

// OAuth defines a controller which handles the incoming request that it contains the giving "secret"
// within it's data.
type OAuth struct {
	Auth    *auth.Auth
	Options []oauth2.AuthCodeOption
	Log     sink.Sink
}

// Redirect attempts to redirect incoming request with the OAuth URL from the supplied OAuth
// structure and uses the giving secret state to generate the URL to redirect to.
func (u *OAuth) Redirect(secret string, w http.ResponseWriter, r *http.Request) (string, error) {
	defer u.Log.Emit(sinks.Info("Redirect Request to OAuth.URL").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"path":   r.URL.Path,
	}).Trace("OAuth.Redirect").End())

	return u.Auth.LoginURL(secret, u.Options...), nil
}

// Validate that the giving SecretCode matches the incoming value of the request else returns an error.
func (u *OAuth) Validate(secret string, w http.ResponseWriter, r *http.Request) error {
	defer u.Log.Emit(sinks.Info("Validated OAuth Secret in Request").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"path":   r.URL.Path,
	}).Trace("OAuth.Validate").End())

	stateSecret := r.FormValue("state")
	if stateSecret != secret {
		err := errors.New("Invalid OAuth secret")
		u.Log.Emit(sinks.Error("OAuth State Fails to match: %+q", err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}))

		return err
	}

	return nil
}
