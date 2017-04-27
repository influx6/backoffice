package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/influx6/backoffice/handlers"
	"github.com/influx6/backoffice/models/session"
	"github.com/influx6/backoffice/utils"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// Sessions exposes a central handle for which requests are served to all requests.
type Sessions struct {
	handlers.Sessions
	Users handlers.Users
}

// Get handles receiving requests to get a sessions from the db.
/* Service API
	HTTP Method: GET
	Request:
		Path: /admin/sessions/:user_id
		Body: None

   Response: (Success, 200)
	Body:
		{
			"user_id":"",
			"public_id":"",
			"expires":"",
			"token":"",
		}

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (s Sessions) Get(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer s.Log.Emit(sinks.Info("Get Existing Session").WithFields(sink.Fields{
		"remote":  r.RemoteAddr,
		"params":  params,
		"path":    r.URL.Path,
		"user_id": params["user_id"],
	}).Trace("Sessions.Create").End())

	userID, ok := params["user_id"]
	if !ok {
		err := errors.New("Expected Session `public_id` as param")
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": params["user_id"],
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	nu, err := s.Sessions.Get(userID)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": params["user_id"],
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user session", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(nu.Fields()); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": params["user_id"],
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// GetAll handles receiving requests to get all sessions from the db.
/* Service API
	HTTP Method: GET
	Request:
		Path: /admin/sessions/
		Body: None

   Response: (Success, 200)
	Body:
		{
			page: 1,
			total: 100,
			responsePerPage: 24,
			records: [{
				"user_id":"",
				"public_id":"",
				"expires":"",
				"token":"",
			}]
		}

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (s Sessions) GetAll(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer s.Log.Emit(sinks.Info("Create New Session").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Sessions.Create").End())

	responsePerPage, _ := strconv.Atoi(params[ResponsePerPageName])
	page, _ := strconv.Atoi(params[PerPageName])

	nus, err := s.Sessions.GetAll(page, responsePerPage)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve sessions", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(nus); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// Login handles receiving requests to create a new session for a user from the server.
/* Service API
	HTTP Method: POST
	Request:
		Path: /sessions/login
		Body:
			{
				"email": "",
				"password": ""
			}

   Response: (Success, 201)
		Body:
			{
				"type":"Bearer",
				"expires":"",
				"token":"",
			}

   Response: (Failure, 500)
	Body:
		`{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (s Sessions) Login(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer s.Log.Emit(sinks.Info("Create New Session").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Sessions.Login").End())

	var nw session.NewSession

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	existingUser, err := s.Users.GetByEmail(nw.Email)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":       r.URL.Path,
			"remote":     r.RemoteAddr,
			"params":     params,
			"user_email": nw.Email,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to find user with email", err)
		return
	}

	if err := existingUser.Authenticate(nw.Password); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":       r.URL.Path,
			"remote":     r.RemoteAddr,
			"params":     params,
			"user_email": nw.Email,
		}))

		utils.WriteErrorMessage(w, http.StatusUnauthorized, "Invalid Credentials: Failed to authenticate user", err)
		return
	}

	newSession, err := s.Sessions.Create(existingUser)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to save new user", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newSession.SessionFields()); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// LogoutWithJSON handles receiving requests to end a user session from the server.
/* Service API
	HTTP Method: DELETE
	Request:
		Path: /sessions/logout/
		Body:
			{
				"user_id": "",
				"token": ""
			}

   Response: (Success, 201)
		Body: None

   Response: (Failure, 500)
	Body:
		`{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (s Sessions) LogoutWithJSON(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer s.Log.Emit(sinks.Info("Delete Existing Session").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Sessions.Logout").End())

	var nw session.EndSession

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": nw.UserID,
		}))

		utils.WriteErrorMessage(w, http.StatusBadRequest, "Failed to read body", err)
		return
	}

	nus, err := s.Sessions.Get(nw.UserID)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": nw.UserID,
		}))
		utils.WriteErrorMessage(w, http.StatusUnauthorized, "Failed to retrieve user's session", err)
		return
	}

	if !nus.ValidateToken(nw.Token) {
		err := errors.New("Invalid User session tokens")
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": nw.UserID,
		}))

		utils.WriteErrorMessage(w, http.StatusUnauthorized, "Failed to validate user's session", err)
		return
	}

	if err := s.Sessions.Delete(nw.UserID); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": nw.UserID,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Logout handles receiving requests to end a user session from the server.
/* Service API
	HTTP Method: DELETE
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}
	Request:
		Path: /sessions/logout/
		Body: None

   Response: (Success, 201)
		Body: None

   Response: (Failure, 500)
	Body:
		`{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (s Sessions) Logout(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer s.Log.Emit(sinks.Info("Delete Existing Session").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Sessions.Logout").End())

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		err := errors.New("Invalid Request: No session currenntly")
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusBadRequest, "Failed to authorization header", err)
		return
	}

	// Retrieve authorization header.
	authType, token, err := utils.ParseAuthorization(authorization)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
			"path":          r.URL.Path,
			"remote":        r.RemoteAddr,
			"params":        params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed decode authorization", err)
		return
	}

	if authType != "Bearer" {
		err := errors.New("Only `Bearer` Authorization supported")
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
			"path":          r.URL.Path,
			"remote":        r.RemoteAddr,
			"params":        params,
		}))

		utils.WriteErrorMessage(w, http.StatusUnauthorized, "Failed decode authorization", err)
		return
	}

	// Retrieve Authorization UserID and Token.
	sessionUserID, sessionToken, err := session.ParseToken(token)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
			"path":          r.URL.Path,
			"remote":        r.RemoteAddr,
			"params":        params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed decode authorization", err)
		return
	}

	nus, err := s.Sessions.Get(sessionUserID)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
			"path":          r.URL.Path,
			"remote":        r.RemoteAddr,
			"params":        params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user's session", err)
		return
	}

	if !nus.ValidateToken(sessionToken) {
		err := errors.New("Invalid User session tokens")
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusUnauthorized, "Failed to validate user's session", err)
		return
	}

	if err := s.Sessions.Delete(nus.PublicID); err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"authorization": authorization,
			"path":          r.URL.Path,
			"remote":        r.RemoteAddr,
			"params":        params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
