package handlers

import (
	"errors"
	"time"

	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/models/session"
	"github.com/influx6/backoffice/utils"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// DefferedBearerAuth returns a new function which can be used to generate a new copy of the
// BearerAuth.
func DefferedBearerAuth(log sink.Sink, dbr db.DB) func(db.TableIdentity, db.TableIdentity, db.TableIdentity, time.Duration) BearerAuth {
	return func(ut db.TableIdentity, pt db.TableIdentity, st db.TableIdentity, expiry time.Duration) BearerAuth {
		users := UsersFactory(log, dbr, ut, pt)
		sessions := SessionsFactory(log, dbr, expiry, st)
		return BearerAuth{
			Users:    users,
			Sessions: sessions,
		}
	}
}

// BearerAuthFactory returns a function which returns a given can be used to generate a
// new Users instance to make request with.
func BearerAuthFactory(log sink.Sink, dbr db.DB, ut db.TableIdentity, pt, st db.TableIdentity, expiry time.Duration) BearerAuth {
	users := UsersFactory(log, dbr, ut, pt)
	sessions := SessionsFactory(log, dbr, expiry, st)
	return BearerAuth{
		Users:    users,
		Sessions: sessions,
	}
}

// BearerAuth defines an handler which provides authorization handling for
// a request, needing user authentication.
type BearerAuth struct {
	Users
	Sessions Sessions
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
func (u BearerAuth) CheckAuthorization(authorization string) error {
	defer u.Log.Emit(sinks.Info("Authenticate Authorization").WithFields(sink.Fields{
		"authorization": authorization,
	}).Trace("Auth.CheckAuthorization").End())

	// Retrieve authorization header.
	authType, token, err := utils.ParseAuthorization(authorization)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	if authType != "Bearer" {
		err := errors.New("Only `Bearer` Authorization supported")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	// Retrieve Authorization UserID and Token.
	sessionUserID, sessionToken, err := session.ParseToken(token)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	// Ensure user does exists.
	if _, err := u.Users.Get(sessionUserID); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	// Retrieve user session record.
	userSession, err := u.Sessions.Get(sessionUserID)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	// if session token does not match UserSession, probably faked request or messed up old session.
	if !userSession.ValidateToken(sessionToken) {
		err := errors.New("Invalid user session's token")

		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	// If session has expired, then we fail the request.
	if userSession.Expired() {
		err := errors.New("User session has expired")

		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
		}))

		return err
	}

	return nil
}
