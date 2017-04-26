package session

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	tableName = "sessions"

	// UniqueIndex defines the unique index name used by the models db for model query optimization.
	UniqueIndex = "user_id"

	// UniqueIndexField defines the unique index field used by the model in it's field.
	UniqueIndexField = "user_public_id"
)

// NewSession defines the set of data received to create a new user's session.
type NewSession struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// EndSession defines the set of data received to end a user's session.
type EndSession struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// Session defines a struct which holds the the details of a giving user session.
type Session struct {
	UserID   string    `json:"user_id"`
	PublicID string    `json:"public_id"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires"`
}

// New returns a new instance of a session.
func New(userID string, expiration time.Time) *Session {
	return &Session{
		UserID:   userID,
		PublicID: uuid.NewV4().String(),
		Token:    uuid.NewV4().String(),
		Expires:  expiration,
	}
}

// ValidateToken validates the provide base64 encoded token, that it matches the
// expected token value with that of the session.
func (u Session) ValidateToken(token string) bool {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	// Attempt to get the session token split which has the userid:session_token.
	sessionToken := strings.Split(string(decoded), ":")
	if len(sessionToken) != 2 {
		return false
	}

	if string(sessionToken[1]) != u.Token {
		return false
	}

	return true
}

// SessionToken returns the Session.Token has a base64 encoded string.
// It returns a base64 encoded version where it contains the UserID:SessionToken.
func (u Session) SessionToken() string {
	sessionToken := fmt.Sprintf("%s:%s", u.UserID, u.Token)
	return base64.StdEncoding.EncodeToString([]byte(sessionToken))
}

// Table returns the given table which the given struct corresponds to.
func (u Session) Table() string {
	return tableName
}

// SessionFields returns a map representing the user session.
func (u Session) SessionFields() map[string]interface{} {
	return map[string]interface{}{
		"type":    "Bearer",
		"token":   u.SessionToken(),
		"expires": u.Expires.Format(time.RFC3339),
	}
}

// Fields returns a map representing the data of the session.
func (u Session) Fields() map[string]interface{} {
	return map[string]interface{}{
		"user_id":   u.UserID,
		"token":     u.Token,
		"public_id": u.PublicID,
		"expires":   u.Expires.Format(time.RFC3339),
	}
}

// Expired returns true/false if the the given session is expired.
func (u *Session) Expired() bool {
	return time.Now().After(u.Expires)
}

// WithFields attempts to syncing the giving data within the provided
// map into it's own fields.
func (u *Session) WithFields(fields map[string]interface{}) error {
	if user, ok := fields["user_id"].(string); ok {
		u.UserID = user
	} else {
		return errors.New("Expected 'user_id' key")
	}

	if public, ok := fields["public_id"].(string); ok {
		u.PublicID = public
	} else {
		return errors.New("Expected 'public_id' key")
	}

	if token, ok := fields["token"].(string); ok {
		u.Token = token
	} else {
		return errors.New("Expected 'token' key")
	}

	if expires, ok := fields["expires"]; ok && expires != "" {
		switch co := expires.(type) {
		case string:
			t, err := time.Parse(time.RFC3339, co)
			if err != nil {
				return err
			}

			u.Expires = t.UTC()
		case time.Time:
			u.Expires = co.UTC()
		}
	}

	return nil
}

// ParseToken parses the base64 encoded token, which it returns the
// associated userID and session token.
func ParseToken(val string) (userID string, token string, err error) {
	var decoded []byte

	decoded, err = base64.StdEncoding.DecodeString(val)
	if err != nil {
		return
	}

	// Attempt to get the session token split which has the userid:session_token.
	sessionToken := strings.Split(string(decoded), ":")
	if len(sessionToken) != 2 {
		err = errors.New("Invalid SessionToken: Token must be UserID:Token  format")
		return
	}

	userID = sessionToken[0]
	token = sessionToken[1]

	return
}
