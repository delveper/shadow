package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/3d0c/gmf"
	"mccoy.space/g/vorbis"
)

func ConvertVoice(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-vn", "-ar", "44100", "-ac", "2", "-ab", "192k", "-f", "mp3", outputPath)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("converting ogg to mp3: %w", err)
	}

	return nil
}

func draft() ([]float32, error) {
	ctx, err := gmf.NewInputCtx("input.ogg")
	if err != nil {
		return nil, err
	}
	_ = ctx
	return nil, nil
}

func decodeOgg(r io.ReadCloser) ([]float32, error) {
	defer func() { _ = r.Close() }()

	ogg, err := vorbis.New(r)
	if err != nil {
		return nil, fmt.Errorf("opening ogg: %w", err)
	}

	data, err := ogg.Decode()
	if err != nil {
		return nil, fmt.Errorf("decoding from ogg: %w", err)
	}

	return data, nil
}

func main() {
	draft()
	// if err := ConvertVoice("input.ogg", "output.mp3"); err != nil {
	// 	log.Fatal(err)
	// }
}
