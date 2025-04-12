package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
)

func generateSessionID() string {
	bytes := make([]byte, 2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (app *application) testHandler(w http.ResponseWriter, r *http.Request) {
	folderId := generateSessionID()
	basepath := "./videos/session-" + folderId
	err := os.MkdirAll(basepath, 0755)
	if err != nil {
		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	subfolder := []string{"hls", "webm"}
	for _, folder := range subfolder {
		path := basepath + "/" + folder
		err := os.MkdirAll(path, 0755)
		if err != nil {
			http.Error(w, "error creating sub folder"+err.Error(), http.StatusInternalServerError)
			return
		}

	}
	w.WriteHeader((http.StatusOK))
	fmt.Println(w, "folders created successfully", folderId)

}
