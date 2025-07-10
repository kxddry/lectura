package uploaded

import "io"

// BrokerRecord MUST HAVE
// STRICT TYPING
type BrokerRecord struct {
	Extension  string `json:"extension"`   // The original extension
	FileName   string `json:"file_name"`   // the original file name (example.mp4) WITHOUT extension
	FileID     string `json:"file_id"`     // generated UUID for the file name
	FileType   string `json:"file_type"`   // Mime Type
	FileSize   int64  `json:"file_size"`   // File Size in Bytes
	WavSize    int64  `json:"wav_size"`    // the file size of Wav File (matches FileSize if input was .wav)
	UploadedAt int64  `json:"uploaded_at"` // time uploaded
}

type FileConfig struct {
	Extension string        // extension of the file
	FileName  string        // the ORIGINAL file name (example.mp4) with extension
	FileID    string        // the UUID generated for the file
	File      io.ReadCloser // an interface for reading the contents
	FileSize  int64         // size of the file in bytes
	Bucket    string        // bucket name for the file to be stored at
	FileType  string        // mime type of the file
}
