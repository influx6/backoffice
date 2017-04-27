package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/influx6/backoffice/handlers"
	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/backoffice/utils"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// Users exposes a central handle for which requests are served to all requests.
type Users struct {
	handlers.Users
}

// GetLimited handles receiving requests to get a user from the db but returns a limited view of the user data.
// This is suited for when needing to respond to requests from non-authorized requests or wishing to exclude
// security based fields (hash, private_id) from the response.
/* Service API
	HTTP Method: GET
	Request:
		Path: /users/:user_id
		Body: None

   Response: (Success, 200)
	Body:
		{
			"public_id":"",
			"email":"",
		}

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (u Users) GetLimited(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Get Existing User").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Users.Get").End())

	publicID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected User `public_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	nu, err := u.Users.Get(publicID)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user", err)
		return
	}

	if err := json.NewEncoder(w).Encode(nu.SafeFields()); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// Get handles receiving requests to get a users from the db.
/* Service API
	HTTP Method: GET
	Request:
		Path: /admin/users/:user_id
		Body: None

   Response: (Success, 200)
	Body:
		{
			"public_id":"",
			"private_id":"",
			"hash":"",
			"email":"",
		}

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (u Users) Get(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Get Existing User").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Users.Get").End())

	publicID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected User `public_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	nu, err := u.Users.Get(publicID)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(nu.Fields()); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// GetAll handles receiving requests to get all users from the db.
/* Service API
	HTTP Method: GET
	Request:
		Path: /admin/users/
		Body: None

   Response: (Success, 200)
	Body:
		{
			page: 1,
			total: 100,
			responsePerPage: 24,
			records: [{
				"public_id":"",
				"private_id":"",
				"hash":"",
				"email":"",
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
func (u Users) GetAll(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Create New User").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Users.GetAll").End())

	responsePerPage, _ := strconv.Atoi(params[ResponsePerPage])
	page, _ := strconv.Atoi(params[Page])

	nus, err := u.Users.GetAll(page, responsePerPage)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve users", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(nus); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// Create handles receiving requests to create a user from the server.
/* Service API
	HTTP Method: POST
	Request:
		Path: /users/
		Body:
		{
			"password":"",
			"email":"",
		}

   Response: (Success, 200)
	Body:
		{
			"public_id":"",
			"email":"",
			"profile":"optional",
		}

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (u Users) Create(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Create New User").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Users.Create").End())

	var nw user.NewUser

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	newUser, err := u.Users.Create(nw)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to save new user", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newUser.SafeFields()); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// UpdatePassword handles receiving requests to update a user identified by it's public_id.
/* Service API
	HTTP Method: PUT
	Request:
		Path: /users/password/:user_id
		Body: None
			{
				"public_id":"",
				"password":"",
				"password_confirmation":"",
			}

   Response: (Success, 201)
		Body: None

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (u Users) UpdatePassword(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Update User Password").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Users.UpdatePassword").End())

	publicID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected User `public_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	var nw user.UpdateUserPassword

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	if nw.PublicID != publicID {
		err := errors.New("JSON User.PublicID does not match update param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to connect to database", err)
		return
	}

	if err := u.Users.UpdatePassword(nw); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to update user password: %+q", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Update handles receiving requests to update a user identified by it's public_id.
/* Service API
	HTTP Method: PUT
	Request:
		Path: /users/:user_id
		Body: None

   Response: (Success, 201)
	Body:
		{
			"public_id":"",
			"email":"",
		}

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (u Users) Update(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Update User").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Users.Update").End())

	publicID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected User `public_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	var nw user.UpdateUser

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	if nw.PublicID != publicID {
		err := errors.New("JSON User.PublicID does not match update param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to connect to database", err)
		return
	}

	if err := u.Users.Update(nw); err != nil {
		err := errors.New("Failed to update user details")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to connect to database", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles receiving requests to removes a user from the server.
/* Service API
	HTTP Method: DELETE
	Request:
		Path: /users/:user_id
		Body: None

   Response: (Success, 201)
		Body: None

   Response: (Failure, 500)
	Body:
		{
			"status":"",
			"title":"",
			"message":"",
		}
*/
func (u Users) Delete(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Delete Existing User").WithFields(sink.Fields{
		"remote":  r.RemoteAddr,
		"params":  params,
		"path":    r.URL.Path,
		"user_id": params["user_id"],
	}).Trace("Users.Delete").End())

	userID, ok := params["user_id"]
	if !ok {
		err := errors.New("Expected User `user_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": params["user_id"],
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read param", err)
		return
	}

	if err := u.Users.Delete(userID); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":    r.URL.Path,
			"remote":  r.RemoteAddr,
			"params":  params,
			"user_id": params["user_id"],
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
