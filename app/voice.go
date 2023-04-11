package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type TempFile struct{ *os.File }

func NewTempFile(name string) (*TempFile, error) {
	file, err := os.CreateTemp("tmp", name)
	if err != nil {
		return nil, fmt.Errorf("creating temp file %q: %w", name, err)
	}

	return &TempFile{
		File: file,
	}, nil
}

func (t *TempFile) Close() error {
	if err := t.File.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Remove(t.Name()); err != nil {
		return fmt.Errorf("removing temp file: %w", err)
	}

	return nil
}

func Convert(input []byte) ([]byte, error) {
	ogg, err := NewTempFile("voice.ogg")
	if err != nil {
		return nil, fmt.Errorf("creating temp ogg file: %w", err)
	}

	defer func() { _ = ogg.Close() }()

	mp3, err := NewTempFile("voice.mp3")
	if err != nil {
		return nil, fmt.Errorf("creating temp mp3 file: %w", err)
	}

	defer func() { _ = mp3.Close() }()

	if _, err := ogg.Write(input); err != nil {
		return nil, fmt.Errorf("writing ogg file: %w", err)
	}

	if err := convertOGGtoMP3(ogg.Name(), mp3.Name()); err != nil {
		return nil, fmt.Errorf("converting ogg to mp3: %w", err)
	}

	buf := new(bytes.Buffer)

	if _, err := mp3.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("resetting mp3 file pointer: %w", err)
	}

	if _, err := buf.ReadFrom(mp3); err != nil {
		return nil, fmt.Errorf("reading mp3 file: %w", err)
	}

	return buf.Bytes(), nil
}

func convertOGGtoMP3(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-vn", "-ar", "44100", "-ac", "2", "-ab", "192k", "-f", "mp3", outputPath)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("converting ogg to mp3: %w", err)
	}

	return nil
}

func ConvertOGGToPCM([]byte) ([]byte, error) {
	return nil, nil
}
