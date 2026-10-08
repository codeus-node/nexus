package nexus

import "github.com/codeus-node/fail"

type HttpError struct {
	Status int
	Error  error
}

func NewHttpError(err fail.CustomError, status int) *HttpError {
	return &HttpError{
		Status: status,
		Error:  err,
	}
}
