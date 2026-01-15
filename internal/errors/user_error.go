package errors

import "net/http"

func EmailAlreadyExists() *AppError {
	return &AppError{
		Code:    "EMAIL_ALREADY_EXISTS",
		Message: "email already registered",
		Status:  http.StatusConflict,
	}
}

func UsernameAlreadyExists() *AppError {
	return &AppError{
		Code:    "USERNAME_ALREADY_EXISTS",
		Message: "username already taken",
		Status:  http.StatusConflict,
	}
}

func InternalServerError() *AppError {
	return &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "something went wrong",
		Status:  http.StatusInternalServerError,
	}
}
