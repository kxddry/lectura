package frontend

type File struct {
	UUID     string `json:"uuid,omitempty"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
