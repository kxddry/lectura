package uploaded

import "io"

type BrokerRecord struct {
	FileName   string `json:"file_name"`
	FileID     string `json:"file_id"`
	FileType   string `json:"file_type"`
	FileSize   int64  `json:"file_size"`
	WavSize    int64  `json:"wav_size"`
	UploadedAt int64  `json:"uploaded_at"`
	S3URL      string `json:"s3_url"`
	WavURL     string `json:"wav_url"`
}

type FileConfig struct {
	Filename string
	FileID   string
	File     io.ReadCloser
	Size     int64
	Bucket   string
	MType    string
}
