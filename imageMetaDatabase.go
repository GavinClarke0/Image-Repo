package main

import (
	"encoding/json"
	"github.com/cockroachdb/pebble"
)

var imageDb *pebble.DB

func OpenImageDb() {

	var err error
	imageDb, err = pebble.Open("image", &pebble.Options{})

	if err != nil {
		panic(err)
	}
}

func GetImageData(key string) (*ImageMeta, error) {

	var meta ImageMeta

	value, _, err := imageDb.Get([]byte(key))

	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(value, &meta)

	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func PutImageData(imageData ImageMeta) error {

	imageKey := imageData.Id

	userBytes, err := json.Marshal(imageData)

	if err != nil {
		return err
	}

	err = imageDb.Set([]byte(imageKey), userBytes, pebble.Sync)

	if err != nil {
		return err
	}
	return nil
}

func GetImagesData(key string) ([]ImageMeta, error) {

	var imagesMeta []ImageMeta

	// determine non-inclusive upper bound for key value iterator
	keyLength := len(key)
	upperBound := key[:keyLength-1] + string(byte(255))

	var iterOptions = &pebble.IterOptions{
		LowerBound: []byte(key),
		UpperBound: []byte(upperBound),
	}

	iter := imageDb.NewIter(iterOptions)

	// include logic to ensure key is valid
	for iter.First(); iter.Valid(); iter.Next() {
		var currentImageMeta ImageMeta

		err := json.Unmarshal(iter.Value(), &currentImageMeta)
		if err != nil {
			return nil, err
		}
		imagesMeta = append(imagesMeta, currentImageMeta)
		// Only keys beginning with "prefix" will be visited.
	}

	return imagesMeta, nil
}

// deletes image meta data from key value store
func DeleteImageMeta(key string) error {
	err := imageDb.Delete([]byte(key), pebble.Sync)

	return err
}
