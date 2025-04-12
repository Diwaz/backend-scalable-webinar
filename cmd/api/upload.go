package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func generateSessionID() string {
	bytes := make([]byte, 2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
func (app *application) handleUpload(w http.ResponseWriter, r *http.Request) {

	// folderID := generateSessionID()
	// basepath := "./"+ folderID

	// only allow POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form data (limit to 100mb)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Failed to parse Multipart form LIMIT EXCEED"+err.Error(), http.StatusBadRequest)
		return
	}

	// retrive file from form data

	file, handler, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "Failed to retrive file"+err.Error(), http.StatusBadRequest)
	}
	defer file.Close()

	// Create uploads directory if not exists
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			http.Error(w, "Failed to create upload directory"+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Generate unique file name with timestamp

	timestap := strconv.FormatInt(time.Now().UnixNano(), 10)
	filename := fmt.Sprintf("chunk-%s.webm", timestap)
	filepath := filepath.Join(uploadDir, filename)

	//create a destination file
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	//copy uploded file to the destination

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file"+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Saved chunk : %s (%s) \n", filename, handler.Header.Get("Content-Type"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chunk uploaded successfully"))

}
