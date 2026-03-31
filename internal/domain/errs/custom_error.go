package errs

//Структуры кастомных ошибок - Доменные ошибки
//Доменные ошибки отдельно от транспортных

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
