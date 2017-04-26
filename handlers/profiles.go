package handlers

import (
	"errors"

	"github.com/influx6/backoffice/db"
	"github.com/influx6/backoffice/models/profile"
	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/faux/sink"
	"github.com/influx6/faux/sink/sinks"
)

// Profiles defines a handler which provides profile related methods.
type Profiles struct {
	DB            db.DB
	Log           sink.Sink
	TableIdentity db.TableIdentity
}

// Create adds a new profile for the specified profile.
func (p Profiles) Create(nu *user.User, np *profile.NewProfile) (*profile.Profile, error) {
	defer p.Log.Emit(sinks.Info("Create New Profile").WithFields(sink.Fields{
		"user_email": nu.Email,
		"user_id":    nu.PublicID,
	}).Trace("Profiles.Create").End())

	if np != nil && np.UserID != nu.PublicID {
		err := errors.New("Invalid NewProfile.UserID: Does not match given user")
		p.Log.Emit(sinks.Error(err).WithFields(sink.Fields{"user_email": nu.Email, "user_id": nu.PublicID}))
		return nil, err
	}

	var newProfile profile.Profile
	profileSeen := true

	// Attempt to retrieve profile from db if we still have an outstanding non-expired profile.
	if err := p.DB.Get(p.Log, p.TableIdentity, &newProfile, profile.UniqueIndex, nu.PublicID); err != nil {
		p.Log.Emit(sinks.Error("Failed to retrieve profile: %+q", err).WithFields(sink.Fields{"user_email": nu.Email, "user_id": nu.PublicID}))
		profileSeen = false
	}

	p.Log.Emit(sinks.Info("New Profile").WithFields(sink.Fields{
		"user_email":   nu.Email,
		"user_id":      nu.PublicID,
		"profile_seen": profileSeen,
	}))

	if profileSeen {
		return &newProfile, nil
	}

	newProfile = *profile.New(nu.PublicID)

	// If the provide data provided is not null, then add changes.
	if np != nil {
		newProfile.Address = np.Address
		newProfile.FirstName = np.FirstName
		newProfile.LastName = np.LastName
	}

	if err := p.DB.Save(p.Log, p.TableIdentity, &newProfile); err != nil {
		p.Log.Emit(sinks.Error("Failed to save profile: %+q", err).WithFields(sink.Fields{"user_email": nu.Email, "user_id": nu.PublicID}))
		return nil, err
	}

	p.Log.Emit(sinks.Info("New Profile Saved").WithFields(sink.Fields{
		"user_email":   nu.Email,
		"user_id":      nu.PublicID,
		"profile_seen": profileSeen,
		"profile":      newProfile,
	}))

	return &newProfile, nil
}

// ProfileRecords defines a struct which returns the total fields and page details
// used in retrieving the records.
type ProfileRecords struct {
	Total         int               `json:"total"`
	Page          int               `json:"page"`
	ResponserPage int               `json:"responserPerPage"`
	Records       []profile.Profile `json:"records"`
}

// GetAll handles receiving requests to retrieve all profile from the database.
func (p Profiles) GetAll(page, responsePerPage int) (ProfileRecords, error) {
	defer p.Log.Emit(sinks.Info("Get Existing User").WithFields(sink.Fields{
		"page":            page,
		"responsePerPage": responsePerPage,
	}).Trace("handlers.Users.Create").End())

	records, realTotalRecords, err := p.DB.GetAllPerPage(p.Log, p.TableIdentity, "asc", "public_id", page, responsePerPage)
	if err != nil {
		p.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"page":            page,
			"responsePerPage": responsePerPage,
		}))
		return ProfileRecords{}, err
	}

	var profileRecords []profile.Profile

	for _, record := range records {
		var nw profile.Profile

		if err := nw.WithFields(record); err != nil {
			p.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
				"page":            page,
				"responsePerPage": responsePerPage,
			}))
			return ProfileRecords{}, err
		}

		profileRecords = append(profileRecords, nw)
	}

	return ProfileRecords{
		Page:          page,
		Total:         realTotalRecords,
		Records:       profileRecords,
		ResponserPage: responsePerPage,
	}, nil
}

// Get retrieves the profile associated with the giving profile_id.
func (p Profiles) Get(profileID string) (*profile.Profile, error) {
	defer p.Log.Emit(sinks.Info("Get Existing Profile").WithFields(sink.Fields{
		"profile_id": profileID,
	}).Trace("Profiles.Get").End())

	var existingProfile profile.Profile

	// Attempt to retrieve profile from db if we still have an outstanding non-expired profile.
	if err := p.DB.Get(p.Log, p.TableIdentity, &existingProfile, "public_id", profileID); err != nil {
		p.Log.Emit(sinks.Error("Failed to retrieve profile from db: %+q", err).WithFields(sink.Fields{"profile_id": profileID}))
		return nil, err
	}

	return &existingProfile, nil
}

// GetByUser retrieves the profile associated with the giving UserID.
func (p Profiles) GetByUser(userID string) (*profile.Profile, error) {
	defer p.Log.Emit(sinks.Info("Get Existing Profile").WithFields(sink.Fields{
		"user_id": userID,
	}).Trace("Profiles.GetByUser").End())

	var existingProfile profile.Profile

	// Attempt to retrieve profile from db if we still have an outstanding non-expired profile.
	if err := p.DB.Get(p.Log, p.TableIdentity, &existingProfile, profile.UniqueIndex, userID); err != nil {
		p.Log.Emit(sinks.Error("Failed to retrieve profile from db: %+q", err).WithFields(sink.Fields{"user_id": userID}))
		return nil, err
	}

	return &existingProfile, nil
}

// DeleteByUser removes an existing profile from the db for a specified profile.
func (p Profiles) DeleteByUser(userID string) error {
	defer p.Log.Emit(sinks.Info("Delete Existing Profile").WithFields(sink.Fields{
		"user_id": userID,
	}).Trace("Profiles.DeleteByUser").End())

	// Delete this profile
	if err := p.DB.Delete(p.Log, p.TableIdentity, profile.UniqueIndex, userID); err != nil {
		p.Log.Emit(sinks.Error("Failed to delete profile profile from db: %+q", err).WithFields(sink.Fields{"user_id": userID}))
		return err
	}

	return nil
}

// Delete removes an existing profile from the db for a specified profile by its id.
func (p Profiles) Delete(profileID string) error {
	defer p.Log.Emit(sinks.Info("Delete Existing Profile").WithFields(sink.Fields{
		"profile_id": profileID,
	}).Trace("Profiles.Delete").End())

	// Delete this profile
	if err := p.DB.Delete(p.Log, p.TableIdentity, "public_id", profileID); err != nil {
		p.Log.Emit(sinks.Error("Failed to delete profile from db: %+q", err).WithFields(sink.Fields{"profile_id": profileID}))
		return err
	}

	return nil
}

// Update handles receiving requests to update a profile identified by it's public_id.
func (p Profiles) Update(nw profile.UpdateProfile) error {
	defer p.Log.Emit(sinks.Info("Update User").With("profile_id", nw.PublicID).Trace("handlers.Users.Update").End())

	if nw.PublicID == "" {
		err := errors.New("JSON User.PublicID is empty")
		p.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"profile_id": nw.PublicID,
		}))

		return err
	}

	if err := p.DB.Update(p.Log, p.TableIdentity, nw, "public_id"); err != nil {
		p.Log.Emit(sinks.Error(err).WithFields(sink.Fields{
			"profile_id": nw.PublicID,
		}))

		return err
	}

	return nil
}
