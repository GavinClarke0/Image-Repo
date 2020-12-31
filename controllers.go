package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
Controllers cover the main logic and serving of data into http form
*/

func NewUser(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = AddUser(creds.Username, creds.Password, createUniqueId())

	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// utilized from source:
func LogIn(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	user, err := GetUser(creds.Username)

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if err != nil || user.Password != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(24 * 60 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &UserClaims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := map[string]string{"token": tokenString}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	_, err = w.Write(bodyBytes)

	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	token := r.Header.Get("Authorization")

	claims, err := Authenticate(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get guid id for user
	userInfo, err := GetUser(claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get image id from url arguments
	imageId, found := vars["imageId"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imageMeta, err := GetImageData(imageId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// image transformations

	// Case where image is private
	if imageMeta.Owner != userInfo.Id && imageMeta.View == "private" {
		w.WriteHeader(http.StatusUnauthorized)
		responseBody := map[string]string{"error": "private image"}
		respBodyBytes, _ := json.Marshal(responseBody)
		_, _ = w.Write(respBodyBytes)
		return
	}

	imageBytes, err := ReadFile(imageMeta.Path)

	// url query params for length and width in pixels
	length := r.URL.Query().Get("length")
	width := r.URL.Query().Get("width")

	// If both parameters are correct transform image
	if length != "" && width != "" {
		lengthInt, err := strconv.Atoi(length)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		widthInt, err := strconv.Atoi(width)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		imageType := strings.ToLower(imageMeta.Type)
		imageBytes, err = ReshapeImage(imageBytes, imageType, lengthInt, widthInt)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	currentTime := time.Now()
	name := imageMeta.Id + "." + imageMeta.Type

	w.Header().Set("Content-Type", "application/"+imageMeta.Type)
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	http.ServeContent(w, r, name, currentTime, bytes.NewReader(imageBytes))
}

// returns all images associated with a profile, returns as ziped file filled with cropped image samples
func GetUserImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	token := r.Header.Get("Authorization")

	claims, err := Authenticate(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get guid id for user
	userInfo, err := GetUser(claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, found := vars["user"]
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	requestedUserInfo, err := GetUser(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// gets image meta for all images of a user
	imagesMeta, err := GetImagesData(requestedUserInfo.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if imagesMeta == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var fw io.Writer
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// iterate through a users images and retrieve all non private images
	for _, image := range imagesMeta {
		if image.Owner == userInfo.Id || image.View != "private" {

			imageFile, err := os.Open(image.Path)
			imageName := image.Id + "." + image.Type

			if fw, err = zw.Create(imageName); err != nil {
				continue
			}
			if _, err = io.Copy(fw, imageFile); err != nil {
				continue
			}
		}
	}
	// close zip file
	currentTime := time.Now()
	_ = zw.Close()

	//name := strconv.FormatInt(currentTime.UTC().UnixNano(), 10) + "-" + userInfo.Id+".zip"
	name := strconv.FormatInt(currentTime.UTC().UnixNano(), 10) + "-photos.zip"

	w.Header().Set("Content-Type", "application/zip")
	http.ServeContent(w, r, name, currentTime, bytes.NewReader(buf.Bytes()))

}

func UploadImage(w http.ResponseWriter, r *http.Request) {

	var imageData ImageMeta
	token := r.Header.Get("Authorization")

	claims, err := Authenticate(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// read image request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get guid id for user
	userInfo, err := GetUser(claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	visibility := r.URL.Query().Get("visibility")

	contentType := r.Header.Get("Content-Type")
	imageType, err := parseImageHeader(contentType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imageData.Type = strings.ToLower(imageType)
	imageData.Owner = userInfo.Id
	imageData.Id = userInfo.Id + "-" + createUniqueId()
	imageData.Path = "./files/" + imageData.Id + "." + imageData.Type

	if visibility == "private" {
		imageData.View = "private"
	} else {
		imageData.View = "public"
	}

	err = WriteFile(body, imageData.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = PutImageData(imageData)

	if err != nil {
		responseBody := map[string]string{"error": "could not save image"}
		respBodyBytes, _ := json.Marshal(responseBody)
		_, _ = w.Write(respBodyBytes)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	responseMessage := map[string]string{"imageId": imageData.Id}
	messageBytes, err := json.Marshal(responseMessage)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(messageBytes)

}

func UploadImages(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Authorization")
	claims, err := Authenticate(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// read image request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get guid id for user
	userInfo, err := GetUser(claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	visibility := r.URL.Query().Get("visibility")

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Fatal(err)
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {

		var imageData ImageMeta

		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			continue
		}

		imageType, err := parseFileType(zipFile.Name)
		if err != nil {
			imageType = "unknown"
		}

		imageData.Type = strings.ToLower(imageType)
		imageData.Owner = userInfo.Id
		imageData.Id = createUniqueId()
		imageData.Path = "./files/" + imageData.Id + "." + imageData.Type

		if visibility == "private" {
			imageData.View = "private"
		} else {
			imageData.View = "public"
		}

		err = WriteFile(unzippedFileBytes, imageData.Path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = PutImageData(imageData)

		if err != nil {
			continue
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func DeleteImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	token := r.Header.Get("Authorization")

	claims, err := Authenticate(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userInfo, err := GetUser(claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imageId, found := vars["imageId"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imageMeta, err := GetImageData(imageId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if imageMeta.Owner == userInfo.Id {
		w.WriteHeader(http.StatusUnauthorized)
	}

	err = os.Remove(imageMeta.Path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = DeleteImageMeta(imageId)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
