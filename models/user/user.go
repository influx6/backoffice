package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/influx6/backoffice/models/profile"
	uuid "github.com/satori/go.uuid"
)

const (
	hashComplexity = 10
	tableName      = "users"
	timeFormat     = "Mon Jan 2 15:04:05 -0700 MST 2006"
)

// UpdateUserPassword defines the set of data sent when updating a users password.
type UpdateUserPassword struct {
	PublicID        string `json:"public_id"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

//====================================================================================================

// UpdateUser defines the set of data sent when updating a users data.
type UpdateUser struct {
	Email    string `json:"email"`
	PublicID string `json:"public_id"`
}

// Fields returns a map representing the data of the user.
func (u UpdateUser) Fields() map[string]interface{} {
	return map[string]interface{}{
		"email":     u.Email,
		"public_id": u.PublicID,
	}
}

// Table returns the given table which the given struct corresponds to.
func (u UpdateUser) Table() string {
	return tableName
}

//====================================================================================================

// NewUser defines the set of data received to create a new user.
type NewUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User is a type defining the given user related fields for a given.
type User struct {
	Email     string           `json:"email"`
	PublicID  string           `json:"public_id"`
	PrivateID string           `json:"private_id,omitempty"`
	Hash      string           `json:"hash,omitempty"`
	Profile   *profile.Profile `json:"profile,omitempty"`
}

// New returns a new User instance based on the provided data.
func New(nw NewUser) (*User, error) {
	var u User
	u.Email = nw.Email
	u.PublicID = uuid.NewV4().String()
	u.PrivateID = uuid.NewV4().String()

	u.ChangePassword(nw.Password)

	return &u, nil
}

// Authenticate attempts to authenticate the giving password to the provided user.
func (u User) Authenticate(password string) error {
	pass := []byte(u.PrivateID + ":" + password)
	return bcrypt.CompareHashAndPassword([]byte(u.Hash), pass)
}

// Table returns the given table which the given struct corresponds to.
func (u User) Table() string {
	return tableName
}

// SafeFields returns a map representing the data of the user with important
// security fields removed.
func (u User) SafeFields() map[string]interface{} {
	fields := u.Fields()

	delete(fields, "hash")
	delete(fields, "private_id")

	return fields
}

// Fields returns a map representing the data of the user.
func (u User) Fields() map[string]interface{} {
	fields := map[string]interface{}{
		"hash":       u.Hash,
		"email":      u.Email,
		"private_id": u.PrivateID,
		"public_id":  u.PublicID,
	}

	if u.Profile != nil {
		fields["profile"] = u.Profile.Fields()
	}

	return fields
}

// ChangePassword uses the provided password to set the users password hash.
func (u *User) ChangePassword(password string) error {
	pass := []byte(u.PrivateID + ":" + password)
	hash, err := bcrypt.GenerateFromPassword(pass, hashComplexity)
	if err != nil {
		return err
	}

	u.Hash = string(hash)
	return nil
}

// WithFields attempts to syncing the giving data within the provided
// map into it's own fields.
func (u *User) WithFields(fields map[string]interface{}) error {
	if email, ok := fields["email"].(string); ok {
		u.Email = email
	} else {
		return errors.New("Expected 'email' key")
	}

	if public, ok := fields["public_id"].(string); ok {
		u.PublicID = public
	} else {
		return errors.New("Expected 'public_id' key")
	}

	if private, ok := fields["private_id"].(string); ok {
		u.PrivateID = private
	} else {
		return errors.New("Expected 'private_id' key")
	}

	if hash, ok := fields["hash"].(string); ok {
		u.Hash = hash
	} else {
		return errors.New("Expected 'hash' key")
	}

	return nil
}
