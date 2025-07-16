package uploaded

import "io"

type Record struct {
	UUID   string `json:"uuid"`
	Bucket string `json:"bucket"`
	// Update struct should only be used by the Updater microservice.
	Update struct {
		UserID      uint   `json:"user_id"`      // 1337
		OGFileName  string `json:"og_file_name"` // "example"
		OGExtension string `json:"og_extension"` // .ogg, .wav, .mp4, .mp3, etc.
		Status      int    `json:"status"`       // default = 0 (uploaded); 1 - transcribed; 2 - summarized (ready)
	} `json:"update"`
}

type File interface {
	FullName() string
	Data() io.Reader
	Size() int64
	MimeType() string
}

var AllowedMimeTypes = map[string]string{
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/x-msvideo": ".avi",
	"audio/aac":       ".aac",
	"audio/wav":       ".wav",
	"audio/ogg":       ".ogg",
	"audio/mpeg":      ".mp3",
}

var Extensions = map[string]string{
	".mp4": "video/mp4",
	".mov": "video/quicktime",
	".avi": "video/x-msvideo",
	".aac": "audio/aac",
	".wav": "audio/wav",
	".ogg": "audio/ogg",
	".mp3": "audio/mpeg",
}
