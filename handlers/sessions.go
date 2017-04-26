package handlers

import (
	"time"

	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/models/session"
	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// Sessions defines a handler which provides session related methods.
type Sessions struct {
	DB            db.DB
	Log           sink.Sink
	Expiration    time.Time
	TableIdentity db.TableIdentity
}

// Create adds a new session for the specified user.
func (s Sessions) Create(nu *user.User) (*session.Session, error) {
	defer s.Log.Emit(sinks.Info("Create New Session").WithFields(sink.Fields{
		"user_email": nu.Email,
		"user_id":    nu.PublicID,
	}).Trace("Sessions.Create").End())

	currentTime := time.Now()

	var newSession session.Session

	// Attempt to retrieve session from db if we still have an outstanding non-expired session.
	if err := s.DB.Get(s.Log, s.TableIdentity, &newSession, session.UniqueIndex, nu.PublicID); err == nil {

		// We have an existing session and the time of expiring is still counting, simly return
		if !newSession.Expires.IsZero() && currentTime.Before(newSession.Expires) {
			return &newSession, nil
		}

		// 	If we still have active session, then simply kick it and return safely.
		if newSession.Expires.IsZero() || currentTime.After(newSession.Expires) {

			// Delete this sessions
			if err := s.DB.Delete(s.Log, s.TableIdentity, session.UniqueIndex, nu.PublicID); err != nil {
				s.Log.Emit(sinks.Error("Failed to delete old session: %+q", err).WithFields(sink.Fields{"user_email": nu.Email, "user_id": nu.PublicID}))
				return nil, err
			}
		}
	}

	// Create new session and store session into db.
	newSession = *session.New(nu.PublicID, s.Expiration)

	if err := s.DB.Save(s.Log, s.TableIdentity, &newSession); err != nil {
		s.Log.Emit(sinks.Error("Failed to save new session: %+q", err).WithFields(sink.Fields{"user_email": nu.Email, "user_id": nu.PublicID}))
		return nil, err
	}

	return &newSession, nil
}

// SessionRecords defines a struct which returns the total fields and page details
// used in retrieving the records.
type SessionRecords struct {
	Total         int               `json:"total"`
	Page          int               `json:"page"`
	ResponserPage int               `json:"responserPerPage"`
	Records       []session.Session `json:"records"`
}

// GetAll handles receiving requests to retrieve all user from the database.
func (s Sessions) GetAll(page, responsePerPage int) (SessionRecords, error) {
	defer s.Log.Emit(sinks.Info("Get Existing User").WithFields(sink.Fields{
		"page":            page,
		"responsePerPage": responsePerPage,
	}).Trace("handlers.Users.Create").End())

	records, realTotalRecords, err := s.DB.GetAllPerPage(s.Log, s.TableIdentity, "asc", "public_id", page, responsePerPage)
	if err != nil {
		s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"page":            page,
			"responsePerPage": responsePerPage,
		}))
		return SessionRecords{}, err
	}

	var sessionRecords []session.Session

	for _, record := range records {
		var nw session.Session

		if err := nw.WithFields(record); err != nil {
			s.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
				"page":            page,
				"responsePerPage": responsePerPage,
			}))
			return SessionRecords{}, err
		}

		sessionRecords = append(sessionRecords, nw)
	}

	return SessionRecords{
		Page:          page,
		Total:         realTotalRecords,
		Records:       sessionRecords,
		ResponserPage: responsePerPage,
	}, nil
}

// Get retrieves the session associated with the giving User.
func (s Sessions) Get(userID string) (*session.Session, error) {
	defer s.Log.Emit(sinks.Info("Get Existing Session").WithFields(sink.Fields{
		"user_id": userID,
	}).Trace("Sessions.Get").End())

	var existingSession session.Session

	// Attempt to retrieve session from db if we still have an outstanding non-expired session.
	if err := s.DB.Get(s.Log, s.TableIdentity, &existingSession, session.UniqueIndex, userID); err != nil {
		s.Log.Emit(sinks.Error("Failed to retrieve session from db: %+q", err).WithFields(sink.Fields{"user_id": userID}))
		return nil, err
	}

	return &existingSession, nil
}

// Delete removes an existing session from the db for a specified user.
func (s Sessions) Delete(userID string) error {
	defer s.Log.Emit(sinks.Info("Delete Existing Session").WithFields(sink.Fields{
		"user_id": userID,
	}).Trace("Sessions.Delete").End())

	// Delete this sessions
	if err := s.DB.Delete(s.Log, s.TableIdentity, session.UniqueIndex, userID); err != nil {
		s.Log.Emit(sinks.Error("Failed to delete user session from db: %+q", err).WithFields(sink.Fields{"user_id": userID}))
		return err
	}

	return nil
}
