package models

//PwdReset Model
type PwdReset struct {
	Name            string
	ConfirmationURL string
	Email           string
	ImagesTemplate  string
}

// CreateAccountConfirm struct
type CreateAccountConfirm struct {
	Name            *string
	Email           string
	City            string
	ConfirmationURL string
	ImagesTemplate  string
}
