package requests

type Request interface {
	Validate() bool
	GetValidationError() error
}
