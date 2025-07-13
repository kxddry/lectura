package uploaded

type KafkaRecord struct {
	UUID string `json:"uuid"`

	// Update struct should only be used by the Updater microservice.
	Update struct {
		UserID      int64  `json:"user_id"`      // 1337
		OGFileName  string `json:"og_file_name"` // "example"
		OGExtension string `json:"og_extension"` // .ogg, .wav, .mp4, .mp3, etc.
		Status      int    `json:"status"`       // default = 0 (uploaded); 1 - transcribed; 2 - summarized (ready)
	} `json:"update"`
}
