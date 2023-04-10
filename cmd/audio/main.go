package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/viert/go-lame"
	ogg "mccoy.space/g/ogg"
)

func convertOGGtoMP3(r io.Reader) ([]byte, error) {
	dec := ogg.NewDecoder(r)

	var pages []*ogg.Page
	for {
		page, err := dec.Decode()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decoding ogg: %w", err)
		}
		pages = append(pages, &page)
	}

	var packets [][]byte
	for _, page := range pages {
		packets = append(packets, page.Packets...)
	}

	buffer := new(bytes.Buffer)

	enc := lame.NewEncoder(buffer)
	defer enc.Close()
	defer func() {
		if _, err := enc.Flush(); err != nil {
			log.Println(err)
		}
	}()

	// if err := enc.SetNumChannels(2); err != nil {
	// 	return nil, fmt.Errorf("setting number of channels: %w", err)
	// }
	//
	// if err := enc.SetQuality(9); err != nil {
	// 	return nil, fmt.Errorf("setting quality: %w", err)
	// }
	//
	// if err := enc.SetBrate(192); err != nil {
	// 	return nil, fmt.Errorf("setting bitrate: %w", err)
	// }

	data := make([]byte, 0)
	for _, packet := range packets {
		dataLen := len(data)
		data = append(data, packet...)
		copy(data[dataLen:], packet)
	}

	n, err := enc.Write(data)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}

	if n != len(data) {
		return nil, fmt.Errorf("not all data was encoded")
	}

	return buffer.Bytes(), nil
}

func main() {
	file, err := os.Open("input.ogg")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	arr, err := convertOGGtoMP3(file)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("output.mp3", arr, 0666); err != nil {
		log.Fatal(err)
	}
}
