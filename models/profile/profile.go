package profile

import (
	uuid "github.com/satori/go.uuid"
)

const (
	tableName  = "profiles"
	timeFormat = "Mon Jan 2 15:04:05 -0700 MST 2006"

	// UniqueIndex defines the unique index name used by the models db for model query optimization.
	UniqueIndex = "user_id"

	// UniqueIndexField defines the unique index field used by the model in it's field.
	UniqueIndexField = "user_public_id"
)

//===============================================================================================

// NewProfile defines a struct which contains data for creating a new user profile.
type NewProfile struct {
	Address   string `json:"address"`
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

//===============================================================================================

// UpdateProfile defines a struct which contains data for updating user profile.
type UpdateProfile struct {
	Address   string `json:"address"`
	PublicID  string `json:"public_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Table returns the given table which the given struct corresponds to.
func (UpdateProfile) Table() string {
	return tableName
}

// Fields returns a map representing the data of the session.
func (u UpdateProfile) Fields() map[string]interface{} {
	return map[string]interface{}{
		"address":    u.Address,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
	}
}

//===============================================================================================

// Profile defines a struct which holds the the details of a giving user's profile.
type Profile struct {
	Address   string `json:"address"`
	UserID    string `json:"user_id"`
	PublicID  string `json:"public_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// New returns a new Profile instance using the supplied userid.
func New(userID string) *Profile {
	return &Profile{
		UserID:   userID,
		PublicID: uuid.NewV4().String(),
	}
}

// Table returns the given table which the given struct corresponds to.
func (Profile) Table() string {
	return tableName
}

// WithFields attempts to syncing the giving data within the provided
// map into it's own fields.
func (u *Profile) WithFields(fields map[string]interface{}) error {
	if user, ok := fields["user_id"].(string); ok {
		u.UserID = user
	}

	if public, ok := fields["public_id"].(string); ok {
		u.PublicID = public
	}

	if firstname, ok := fields["first_name"].(string); ok {
		u.FirstName = firstname
	}

	if lastname, ok := fields["last_name"].(string); ok {
		u.LastName = lastname
	}

	if address, ok := fields["address"].(string); ok {
		u.Address = address
	}

	return nil
}

// Fields returns a map representing the data of the session.
func (u *Profile) Fields() map[string]interface{} {
	return map[string]interface{}{
		"address":    u.Address,
		"user_id":    u.UserID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"public_id":  u.PublicID,
	}
}
