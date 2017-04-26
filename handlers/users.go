package handlers

import (
	"errors"

	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// DeferredUsersFactory returns a function which allows easily create a new copy
// of a Users struct.
func DeferredUsersFactory(log sink.Sink, dbr db.DB) func(db.TableIdentity, db.TableIdentity) Users {
	return func(ut db.TableIdentity, pt db.TableIdentity) Users {
		users := Users{
			DB:            dbr,
			Log:           log,
			TableIdentity: ut,
		}

		if pt != nil {
			pm := ProfilesFactory(log, dbr, pt)
			users.Profiles = &pm
		}

		return users
	}
}

// UsersFactory returns a function which returns a given can be used to generate a
// new Users instance to make request with.
func UsersFactory(log sink.Sink, dbr db.DB, usersT db.TableIdentity, profilesT db.TableIdentity) Users {
	users := Users{
		DB:            dbr,
		Log:           log,
		TableIdentity: usersT,
	}

	if profilesT != nil {
		pm := ProfilesFactory(log, dbr, profilesT)
		users.Profiles = &pm
	}

	return users
}

// Users exposes a central handle for which requests are served to all requests.
type Users struct {
	DB            db.DB
	Log           sink.Sink
	Profiles      *Profiles
	TableIdentity db.TableIdentity
}

// Delete handles receiving requests to delete a user from the database.
func (u Users) Delete(id string) error {
	defer u.Log.Emit(sinks.Info("Get Existing User").With("user_id", id).Trace("handlers.Users.Create").End())

	if err := u.DB.Delete(u.Log, u.TableIdentity, "public_id", id); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"public_id": id}))
		return err
	}

	var err error

	// Delete user profile.
	if u.Profiles != nil {
		if err = u.Profiles.DeleteByUser(id); err != nil {
			u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"public_id": id}))
			return err
		}
	}

	return nil
}

// Get handles receiving requests to retrieve a user from the database.
func (u Users) Get(id string) (*user.User, error) {
	defer u.Log.Emit(sinks.Info("Get Existing User").With("user_id", id).Trace("handlers.Users.Create").End())

	var nu user.User

	if err := u.DB.Get(u.Log, u.TableIdentity, &nu, "public_id", id); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"public_id": id}))
		return nil, err
	}

	var err error

	// Get user profile.
	if u.Profiles != nil {
		nu.Profile, err = u.Profiles.GetByUser(nu.PublicID)
		if err != nil {
			u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"public_id": id}))
			return nil, err
		}
	}

	return &nu, nil
}

// GetByEmail handles receiving requests to retrieve a user with user's email from the database.
func (u Users) GetByEmail(email string) (*user.User, error) {
	defer u.Log.Emit(sinks.Info("Get Existing User").With("user_email", email).Trace("handlers.Users.Create").End())

	var nu user.User

	if err := u.DB.Get(u.Log, u.TableIdentity, &nu, "email", email); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"user_email": email}))
		return nil, err
	}

	var err error

	// Get user profile.
	if u.Profiles != nil {
		nu.Profile, err = u.Profiles.GetByUser(nu.PublicID)
		if err != nil {
			u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"user_email": email}))
			return nil, err
		}
	}

	return &nu, nil
}

// UserRecords defines a struct which returns the total fields and page details
// used in retrieving the records.
type UserRecords struct {
	Total         int         `json:"total"`
	Page          int         `json:"page"`
	ResponserPage int         `json:"responserPerPage"`
	Records       []user.User `json:"records"`
}

// GetAll handles receiving requests to retrieve all user from the database.
func (u Users) GetAll(page, responsePerPage int) (UserRecords, error) {
	defer u.Log.Emit(sinks.Info("Get Existing User").WithFields(sink.Fields{
		"page":            page,
		"responsePerPage": responsePerPage,
	}).Trace("handlers.Users.Create").End())

	records, realTotalRecords, err := u.DB.GetAllPerPage(u.Log, u.TableIdentity, "asc", "public_id", page, responsePerPage)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"page":            page,
			"responsePerPage": responsePerPage,
		}))

		return UserRecords{}, err
	}

	var userRecords []user.User

	for _, record := range records {
		var nw user.User

		if err := nw.WithFields(record); err != nil {
			u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
				"page":            page,
				"responsePerPage": responsePerPage,
			}))
			return UserRecords{}, err
		}

		userRecords = append(userRecords, nw)
	}

	return UserRecords{
		Page:          page,
		Total:         realTotalRecords,
		ResponserPage: responsePerPage,
		Records:       userRecords,
	}, nil
}

// Create handles receiving requests to create a user from the server.
func (u Users) Create(nw user.NewUser) (*user.User, error) {
	defer u.Log.Emit(sinks.Info("Create New User").Trace("handlers.Users.Create").End())

	newUser, err := user.New(nw)
	if err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"email": nw.Email}))
		return nil, err
	}

	if err := u.DB.Save(u.Log, u.TableIdentity, newUser); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"email": nw.Email}))
		return nil, err
	}

	// Add user profile.
	if u.Profiles != nil {
		newUser.Profile, err = u.Profiles.Create(newUser, nil)
		if err != nil {
			u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"email": nw.Email}))
			return nil, err
		}
	}

	return newUser, nil
}

// UpdatePassword handles receiving requests to update a user identified by it's public_id.
func (u Users) UpdatePassword(nw user.UpdateUserPassword) error {
	defer u.Log.Emit(sinks.Info("Update User Password").With("user", nw.PublicID).Trace("handlers.Users.UpdatePassword").End())

	if nw.PublicID == "" {
		err := errors.New("JSON UpdateUserPassword.PublicID is empty")

		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
		}))

		return err
	}

	// TODO(influx6): Should we do some password validty checks.
	if nw.Password == "" {
		err := errors.New("JSON UpdateUserPassword.Password is empty")

		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
		}))

		return err
	}

	// TODO(influx6): Do we need to do this here.
	// if nw.Password != nw.PasswordConfirm {
	// 	err := errors.New("Invalid Confirmation Password")
	// 	u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
	//		"user_id":   nw.PublicID,
	// 	}))
	// 	http.Error(w, utils.ErrorMessage(http.StatusInternalServerError, "Failed to connect to database", err), http.StatusInternalServerError)
	// 	return
	// }

	var dbUser user.User

	if err := u.DB.Get(u.Log, u.TableIdentity, &dbUser, "public_id", nw.PublicID); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
		}))

		return err
	}

	if err := dbUser.ChangePassword(nw.Password); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
		}))

		return err
	}

	if err := u.DB.Update(u.Log, u.TableIdentity, &dbUser, "public_id"); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
		}))

		return err
	}

	return nil
}

// Update handles receiving requests to update a user identified by it's public_id.
func (u Users) Update(nw user.UpdateUser) error {
	defer u.Log.Emit(sinks.Info("Update User").With("user", nw.PublicID).Trace("handlers.Users.Update").End())

	if nw.PublicID == "" {
		err := errors.New("JSON User.PublicID is empty")
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
			"email":   nw.Email,
		}))

		return err
	}

	if err := u.DB.Update(u.Log, u.TableIdentity, nw, "public_id"); err != nil {
		u.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"user_id": nw.PublicID,
			"email":   nw.Email,
		}))

		return err
	}

	return nil
}
