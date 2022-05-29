package auth

// ValidationError is an error for when a table entry isn't valid
type ValidationError struct {
	Messages map[string]string
}

// UserNotFoundError is an error for when an user is not found in the database
type UserNotFoundError struct {
	Message string
}

// PwdResetInvalidError is an error for when a password reset id is not found in the database
type PwdResetInvalidError struct {
	Message string
}

func (e ValidationError) Error() (stringy string) {
	for _, v := range e.Messages {
		stringy += v + "\r\n"
	}
	return
}

func (e UserNotFoundError) Error() string {
	return e.Message
}

func (e PwdResetInvalidError) Error() string {
	return e.Message
}
