package app

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

type TempFile struct{ *os.File }

func NewTempFile(name string) (*TempFile, error) {
	file, err := os.Create(path.Join("tmp", name))
	if err != nil {
		return nil, fmt.Errorf("creating temp file %q: %w", name, err)
	}

	tmp := TempFile{File: file}

	return &tmp, nil
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
	ogg, err := NewTempFile("voice.oga")
	if err != nil {
		return nil, fmt.Errorf("creating temp oga file: %w", err)
	}

	defer func() { _ = ogg.Close() }()

	mp3, err := NewTempFile("voice.mp3")
	if err != nil {
		return nil, fmt.Errorf("creating temp mp3 file: %w", err)
	}

	defer func() { _ = mp3.Close() }()

	if n, err := ogg.Write(input); err != nil {
		return nil, fmt.Errorf("writing oga file: %w", err)
	} else {
		log.Println("wrote", n, "bytes to", ogg.Name())
	}

	if err := convertOGGtoMP3(ogg.Name(), mp3.Name()); err != nil {
		return nil, fmt.Errorf("converting oga to mp3: %w", err)
	}

	buf := new(bytes.Buffer)

	if n, err := buf.ReadFrom(mp3); err != nil {
		return nil, fmt.Errorf("reading mp3 file: %w", err)
	} else {
		log.Println("read", n, "bytes from", mp3.Name())
	}

	return buf.Bytes(), nil
}

func convertOGGtoMP3(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-vn", "-ar", "48000", "-ac", "1", "-ab", "36k", "-f", "mp3", outputPath, "-y")
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("converting ogg to mp3: %w", err)
	}

	return nil
}
