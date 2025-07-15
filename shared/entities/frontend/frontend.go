package frontend

type File struct {
	UUID     string `json:"uuid,omitempty"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}
