package main

import "io/ioutil"

func WriteFile(imageData []byte, path string) error {

	// determine what that will  do
	err := ioutil.WriteFile(path, imageData, 0666)
	if err != nil {
		return err
	}
	return nil
}

func ReadFile(path string) ([]byte, error) {

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return dat, nil

}
