package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func main() {

	OpenAccountDb()
	OpenImageDb()

	router := mux.NewRouter()
	router.HandleFunc("/createuser", NewUser).Methods("POST")
	router.HandleFunc("/login", LogIn).Methods("POST")
	router.HandleFunc("/image/getimage/{imageId}", GetImage).Methods("GET")
	router.HandleFunc("/image/getallusermmages/{user}", GetUserImages).Methods("GET")
	router.HandleFunc("/image/putimage", UploadImage).Methods("POST")
	router.HandleFunc("/image/putimages", UploadImages).Methods("POST")
	router.HandleFunc("/image/deleteimage/id", DeleteImage).Methods("DELETE")

	err := http.ListenAndServe(":8333", router)
	if err != nil {
		panic("ERROR: Could not start server")
	}

}
