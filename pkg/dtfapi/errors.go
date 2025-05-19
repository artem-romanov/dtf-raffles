package dtfapi

import (
	"errors"
	"fmt"
)

type DtfErrorV2 struct {
	Message string `json:"message"`
	Err     struct {
		Code int `json:"code"`
	} `json:"error"`
}

func (err DtfErrorV2) Error() string {
	return fmt.Sprintf(`api error: %s (code: %d)`, err.Message, err.Err.Code)
}

type DtfErrorV3 struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var ErrMissingToken = errors.New("Access token is not provided")
var ErrInvalidCredentials = errors.New("Invalid email & password")

func (err DtfErrorV3) Error() string {
	return fmt.Sprintf(`api error: %s (code: %d)`, err.Message, err.Code)
}
