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
	// Prepare paths
	tmpInputPath := "/tmp/" + fc.FileID + fc.Extension
	tmpOutputPath := "/tmp/" + fc.FileID + ".wav"

	// Make sure we clean up both temp files
	defer func() {
		_ = os.Remove(tmpInputPath)
		_ = os.Remove(tmpOutputPath)
	}()

	// Write incoming file to a temp
	inFile, err := os.Create(tmpInputPath)
	if err != nil {
		return uploaded.FileConfig{},
			fmt.Errorf("failed to create temp input file: %w", err)
	}
	// Caller must close the original fc.File; we only close our temp writer
	defer inFile.Close()

	if _, err := io.Copy(inFile, fc.File); err != nil {
		return uploaded.FileConfig{},
			fmt.Errorf("failed to write to temp input file: %w", err)
	}

	// Convert with ffmpeg
	if err := convertFileToWav(tmpInputPath, tmpOutputPath); err != nil {
		return uploaded.FileConfig{}, err
	}

	// Open the converted wav
	outFile, err := os.Open(tmpOutputPath)
	if err != nil {
		return uploaded.FileConfig{},
			fmt.Errorf("failed to open converted wav: %w", err)
	}

	info, err := outFile.Stat()
	if err != nil {
		_ = outFile.Close()
		return uploaded.FileConfig{},
			fmt.Errorf("failed to stat converted wav: %w", err)
	}

	fcc := uploaded.FileConfig{
		Extension: ".wav",
		FileName:  fc.FileName,
		FileID:    fc.FileID,
		File:      outFile,
		FileSize:  info.Size(),
		Bucket:    fc.Bucket,
		FileType:  "audio/wav",
	}
	return fcc, nil
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
		return fmt.Errorf("ffmpeg error: %v", err)
	}
	return nil
}
