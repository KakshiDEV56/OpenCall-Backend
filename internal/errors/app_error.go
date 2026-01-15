package errors

type AppError struct {
	Code    string `json:"code"`    // semantic code
	Message string `json:"message"` // human-readable
	Status  int    `json:"-"`       // HTTP status (hidden from client)
}

func (e *AppError) Error() string {
	return e.Message
}
