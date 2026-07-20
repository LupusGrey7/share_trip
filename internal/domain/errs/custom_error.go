// Custom errors - Domain errors
// Domain errors are separate from transport errors
package errs

type RequestValidationError struct {
	Message string
}

func (err RequestValidationError) Error() string {
	return err.Message
}

type JsonParseValidationError struct {
	Message string
}

func (err JsonParseValidationError) Error() string {
	return err.Message
}
