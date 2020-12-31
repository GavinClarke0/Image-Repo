package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	guuid "github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"strings"
)

func WriteRespone(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func createUniqueId() string {
	id := guuid.New()
	return id.String()
}

func parseImageHeader(s string) (string, error) {

	index := strings.Index(s, "/")
	if index == -1 {
		return s, errors.New("char not found")
	}

	return s[index+1:], nil
}

func parseFileType(s string) (string, error) {

	index := strings.Index(s, ".")
	if index == -1 {
		return s, errors.New("char not found")
	}

	return s[index+1:], nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
