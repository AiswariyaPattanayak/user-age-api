package models

import "errors"

const DateLayout = "2006-01-02"

var (
	ErrNotFound   = errors.New("user not found")
	ErrInvalidDOB = errors.New("dob must be a valid date in the past (YYYY-MM-DD)")
)

// UserRequest is used for both create (POST) and update (PUT).
type UserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	DOB  string `json:"dob" validate:"required,datetime=2006-01-02"`
}

// UserResponse is returned by create and update.
type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
}

// UserWithAgeResponse is returned by get and list.
type UserWithAgeResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
	Age  int    `json:"age"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}
