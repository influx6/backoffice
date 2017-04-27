package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrorMessage returns a string which contains a json value of a
// given error message to be delivered.
func ErrorMessage(status int, header string, err error) string {
	return fmt.Sprintf(`{
		"status": %d,
		"title": %+q,
		"message": %+q,
	}`, status, header, err)
}

// WriteErrorMessage writes the giving error message to the provided writer.
func WriteErrorMessage(w http.ResponseWriter, status int, header string, err error) {
	http.Error(w, ErrorMessage(status, header, err), status)
}

// ParseAuthorization returns the scheme and token of the Authorization string
// if it's valid.
func ParseAuthorization(val string) (authType string, token string, err error) {
	authSplit := strings.SplitN(val, " ", 2)
	if len(authSplit) != 2 {
		err = errors.New("Invalid Authorization: Expected content: `AuthType Token`")
		return
	}

	authType = strings.TrimSpace(authSplit[0])
	token = strings.TrimSpace(authSplit[1])

	return
}
