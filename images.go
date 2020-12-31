package main

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
)

func ReshapeImage(imageFile []byte, imageType string, length int, width int) ([]byte, error) {

	originalImage, _, err := image.Decode(bytes.NewReader(imageFile))

	if err != nil {
		return nil, err
	}

	newImage := resize.Resize(uint(width), uint(length), originalImage, resize.Lanczos3)

	return encodeImage(newImage, imageType)
}

func encodeImage(newImage image.Image, imageType string) ([]byte, error) {

	buf := new(bytes.Buffer)

	switch imageType {
	case "jpeg":
		err := jpeg.Encode(buf, newImage, nil)
		if err != nil {
			return nil, err
		}

		imageBytes := buf.Bytes()
		return imageBytes, nil
	case "png":
		err := png.Encode(buf, newImage)
		if err != nil {
			return nil, err
		}

		imageBytes := buf.Bytes()
		return imageBytes, nil
	default:
		return nil, errors.New("incorrect image type supplied")
	}
}
