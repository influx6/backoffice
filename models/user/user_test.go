package user_test

import (
	"testing"

	"github.com/influx6/backoffice/models/user"
	"github.com/influx6/faux/tests"
)

// TestUserWithField validates the with Field method.
func TestUserWithField(t *testing.T) {
	var nw user.User

	if err := nw.WithFields(map[string]interface{}{
		"email":      "buba@gum.com",
		"hash":       "2332323-23220-Gu34433-23232232",
		"public_id":  "2332323-23220-Gu34433-23232232",
		"private_id": "2332323-23220-Gu34433-23232232",
	}); err != nil {
		tests.Failed("Should have successfully field user with fields: %+q.", err)
	}
	tests.Passed("Should have successfully field user with fields.")

	if nw.Email != "buba@gum.com" {
		tests.Failed("Should have matched expected email on user.")
	}
	tests.Passed("Should have matched expected email on user.")
}

// TestUser validates the methods and returns attached to the user model.
func TestUser(t *testing.T) {
	oldUser, err := user.New(user.NewUser{
		Email:    "bob@guma.com",
		Password: "glow",
	})

	if err != nil {
		tests.Failed("Should have successfully created new user: %+q.", err)
	}
	tests.Passed("Should have successfully created new user.")

	// Validate returned fields to match expected name set.
	{
		fields := oldUser.Fields()

		if _, ok := fields["public_id"]; !ok {
			tests.Failed("Should have a 'public_id' field")
		}
		tests.Passed("Should have a 'public_id' field")

		if _, ok := fields["private_id"]; !ok {
			tests.Failed("Should have a 'private_id' field")
		}
		tests.Passed("Should have a 'private_id' field")

		if _, ok := fields["email"]; !ok {
			tests.Failed("Should have a 'email' field")
		}
		tests.Passed("Should have a 'email' field")

		if _, ok := fields["hash"]; !ok {
			tests.Failed("Should have a 'hash' field")
		}
		tests.Passed("Should have a 'hash' field")

	}

	if err := oldUser.Authenticate("glow"); err != nil {
		tests.Failed("Should have successfully authenticated with provided password: %+q.", err)
	}
	tests.Passed("Should have successfully authenticated with provided password.")
}
