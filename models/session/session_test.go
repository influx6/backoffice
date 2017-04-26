package session_test

import (
	"testing"
	"time"

	"github.com/influx6/backoffice/models/session"
	"github.com/influx6/faux/tests"
)

// TestSessionWithField validates the with Field method.
func TestSessionWithField(t *testing.T) {
	var nw session.Session

	if err := nw.WithFields(map[string]interface{}{
		"token":     "2332323-23220-Gu34433-23232232",
		"user_id":   "2332323-23220-Gu34433-23232232",
		"public_id": "2332323-23220-Gu34433-23232232",
		"expires":   time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		tests.Failed("Should have successfully filled session with fields: %+q.", err)
	}
	tests.Passed("Should have successfully filled session with fields.")

	if nw.Token != "2332323-23220-Gu34433-23232232" {
		tests.Failed("Should have matched expected token on session.")
	}
	tests.Passed("Should have matched expected token on session.")
}

// TestSession validates the methods and returns attached to the session model.
func TestSession(t *testing.T) {
	var se session.Session

	// Validate returned fields to match expected name set.
	fields := se.Fields()

	if _, ok := fields["public_id"]; !ok {
		tests.Failed("Should have a 'public_id' field")
	}
	tests.Passed("Should have a 'public_id' field")

	if _, ok := fields["user_id"]; !ok {
		tests.Failed("Should have a 'user_id' field")
	}
	tests.Passed("Should have a 'user_id' field")

	if _, ok := fields["token"]; !ok {
		tests.Failed("Should have a 'token' field")
	}
	tests.Passed("Should have a 'token' field")

	if _, ok := fields["expires"]; !ok {
		tests.Failed("Should have a 'expires' field")
	}
	tests.Passed("Should have a 'expires' field")
}
