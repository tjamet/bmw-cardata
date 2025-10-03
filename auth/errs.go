package auth

import "fmt"

type AuthError struct {
	StatusCode  int    `json:"status_code,omitempty"`
	Err         string `json:"error,omitempty"`
	Description string `json:"error_description,omitempty"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%d: %s: %s", e.StatusCode, e.Err, e.Description)
}
