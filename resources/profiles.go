package resources

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/influx6/backoffice/handlers"
	"github.com/influx6/backoffice/models/profile"
	"github.com/influx6/backoffice/utils"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// Profiles exposes a central handle for which the API exposes request for profiles.
type Profiles struct {
	handlers.Profiles
	Users    handlers.Users
	Sessions handlers.Sessions
}

// GetForUser handles receiving requests to get a user's profile from the backend.
/* Service API
	HTTP Method: GET
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}

			WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>

	Request:
		Path: /profile/users/:user_id
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
func (u Profiles) GetForUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Get Existing Profile").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Profiles.Get").End())

	// Retrieve UserID from the params.
	userID, ok := params["user_id"]
	if !ok {
		err := errors.New("Expected Profile `UserID` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusBadRequest, "Failed to read `user_id` in params", err)
		return
	}

	nu, err := u.Profiles.GetByUser(userID)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user's profile", err)
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

// Get handles receiving requests to get a users from the db.
/* Service API
	HTTP Method: GET
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}

			WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>

	Request:
		Path: /profiles/:public_id
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
func (u Profiles) Get(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Get Existing Profile").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Profiles.Get").End())

	publicID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected Profile `public_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	nu, err := u.Profiles.Get(publicID)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to retrieve user profile", err)
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
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}

			WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>

	Request:
		Path: /admin/profiles/
		Body: None

   Response: (Success, 200)
	Body:
		{
			page: 1,
			total: 100,
			responsePerPage: 24,
			records: [{
				"first_name":"",
				"last_name":"",
				"user_id":"",
				"profile_id":"",
				"email":"",
				"address":"",
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
func (u Profiles) GetAll(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Create New Profile").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Profiles.GetAll").End())

	responsePerPage, _ := strconv.Atoi(params[ResponsePerPageName])
	page, _ := strconv.Atoi(params[PerPageName])

	nus, err := u.Profiles.GetAll(page, responsePerPage)
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
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}

			WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>

	Request:
		Path: /profiles/
		Body:
			{
				"first_name":"",
				"last_name":"",
				"user_id":"",
				"email":"",
				"address":"",
			}

   Response: (Success, 200)
		Body:
			{
				"first_name":"",
				"last_name":"",
				"user_id":"",
				"public_id":"",
				"email":"",
				"address":"",
			}

   Response: (Failure, 500)
		Body:
			{
				"status":"",
				"title":"",
				"message":"",
			}
*/
func (u Profiles) Create(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Create New Profile").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Profiles.Create").End())

	var nw profile.NewProfile

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

	existingUser, err := u.Users.Get(nw.UserID)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to get user for profile", err)
		return
	}

	newProfile, err := u.Profiles.Create(existingUser, &nw)
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

	if err := json.NewEncoder(w).Encode(newProfile.Fields()); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to return new user data", err)
		return
	}
}

// Update handles receiving requests to update a user identified by it's public_id.
/* Service API
	HTTP Method: PUT
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}

			WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>

	Request:
		Path: /profile/:public_id
		Body:
			{
				"first_name":"",
				"last_name":"",
				"public_id":"",
				"email":"",
				"address":"",
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
func (u Profiles) Update(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Update Profile").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Profiles.Update").End())

	publicID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected Profile `public_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read body", err)
		return
	}

	var nw profile.UpdateProfile

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
		err := errors.New("JSON Profile.PublicID does not match update param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to connect to database", err)
		return
	}

	if err := u.Profiles.Update(nw); err != nil {
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
	Header:
			{
				"Authorization":"Bearer <TOKEN>",
			}

			WHERE: <TOKEN> = <USERID>:<SESSIONTOKEN>

	Request:
		Path: /profile/:public_id
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
func (u Profiles) Delete(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer u.Log.Emit(sinks.Info("Delete Existing Profile").WithFields(sink.Fields{
		"remote": r.RemoteAddr,
		"params": params,
		"path":   r.URL.Path,
	}).Trace("Profiles.Delete").End())

	profileID, ok := params["public_id"]
	if !ok {
		err := errors.New("Expected Profile `profile_id` as param")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))

		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to read param", err)
		return
	}

	if err := u.Profiles.Delete(profileID); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"params": params,
		}))
		utils.WriteErrorMessage(w, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
