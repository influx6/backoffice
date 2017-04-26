package profile_test

import (
	"testing"

	"github.com/influx6/backoffice/models/profile"
	"github.com/influx6/faux/tests"
)

// TestProfileWithField validates the with Field method.
func TestProfileWithField(t *testing.T) {
	var nw profile.Profile

	if err := nw.WithFields(map[string]interface{}{
		"first_name": "Bob",
		"last_name":  "Ged",
		"address":    "No. 20 Toku street, Ala, Lagos.",
		"user_id":    "2332323-23220-Gu34433-23232232",
		"public_id":  "2332323-23220-Gu34433-23232232",
	}); err != nil {
		tests.Failed("Should have successfully filled profile with fields: %+q.", err)
	}
	tests.Passed("Should have successfully filled profile with fields.")

	if nw.UserID != "2332323-23220-Gu34433-23232232" {
		tests.Failed("Should have matched expected UserID on profile.")
	}
	tests.Passed("Should have matched expected UserID on profile.")
}

// TestProfile validates the methods and returns attached to the profile model.
func TestProfile(t *testing.T) {
	var se profile.Profile

	// Validate returned fields to match expected name set.
	fields := se.Fields()

	if _, ok := fields["address"]; !ok {
		tests.Failed("Should have a 'address' field")
	}
	tests.Passed("Should have a 'address' field")

	if _, ok := fields["user_id"]; !ok {
		tests.Failed("Should have a 'user_id' field")
	}
	tests.Passed("Should have a 'user_id' field")

	if _, ok := fields["public_id"]; !ok {
		tests.Failed("Should have a 'public_id' field")
	}
	tests.Passed("Should have a 'public_id' field")

	if _, ok := fields["first_name"]; !ok {
		tests.Failed("Should have a 'first_name' field")
	}
	tests.Passed("Should have a 'first_name' field")

	if _, ok := fields["last_name"]; !ok {
		tests.Failed("Should have a 'last_name' field")
	}
	tests.Passed("Should have a 'last_name' field")
}
