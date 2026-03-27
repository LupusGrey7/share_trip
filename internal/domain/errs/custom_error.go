package errs

//Структуры кастомных ошибок - Доменные ошибки
//Доменные ошибки отдельно от транспортных

type RequestValidationError struct {
	Message string
}

type JsonParseValidationError struct {
	Message string
}

func (err RequestValidationError) Error() string {
	return err.Message
}
func (err JsonParseValidationError) Error() string {
	return err.Message
}
