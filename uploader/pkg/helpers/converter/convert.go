package converter

import (
	"bytes"
	"fmt"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"io"
	"os"
	"os/exec"
)

// ConvertToWav converts a file to wav and returns a FileConfig.
// Always defer closing the fc.File reader after calling the function
func ConvertToWav(fc uploaded.FileConfig) (uploaded.FileConfig, error) {
	fcc := uploaded.FileConfig{}
	tmpInputPath := "/tmp/" + fc.Filename
	tmpOutputPath := "/tmp/" + fc.FileID + ".wav"

	out, err := os.Create(tmpInputPath)
	defer os.Remove(tmpInputPath)
	defer out.Close()

	if err != nil {
		return fcc, fmt.Errorf("%s: %w", "failed to create a temporary file", err)
	}
	if _, err = io.Copy(out, fc.File); err != nil {
		return fcc, fmt.Errorf("%s: %w", "failed to fill the temp file up", err)
	}

	if err = convertFileToWav(tmpInputPath, tmpOutputPath); err != nil {
		return fcc, err
	}

	file, err := os.Open(tmpOutputPath)
	if err != nil {
		return fcc, fmt.Errorf("%s: %w", "failed to open the converted file???", err)
	}
	info, _ := file.Stat()
	fcc = uploaded.FileConfig{
		Filename: fc.FileID + ".wav",
		FileID:   fc.FileID,
		File:     file,
		Size:     info.Size(),
		Bucket:   fc.Bucket,
		MType:    fc.MType,
	}
	return fc, nil
}

func convertFileToWav(inputPath, outPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ar", "16000", // sampling rate
		"-ac", "1", // mono
		"-f", "wav", outPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg error: %v, details: %s", err, stderr.String())
	}
	return nil
}
