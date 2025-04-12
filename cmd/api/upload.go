package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	sessionCounters = make(map[string]uint64)
	mu              sync.Mutex
)

func waitForFileStability(filePath string, interval time.Duration) bool {
	const maxRetries = 5
	var lastSize int64 = -1

	for i := 0; i < maxRetries; i++ {
		info, err := os.Stat(filePath)
		if err != nil {
			log.Println("File stat error:", err)
			return false
		}

		currentSize := info.Size()

		if currentSize == lastSize {
			return true // ✅ File is stable
		}

		lastSize = currentSize
		time.Sleep(interval)
	}

	log.Println("File is not stable after retries")
	return false
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

	// retrive sessionId from form
	sessionID := r.FormValue("sessionId")
	log.Println("sessionid", sessionID)
	if sessionID == "" {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
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

	mu.Lock()
	counter := sessionCounters[sessionID] + 1
	sessionCounters[sessionID] = counter
	mu.Unlock()

	basepath := "./uploads/session-" + sessionID
	err = os.MkdirAll(basepath, 0755)
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
	// Generate unique file name with timestamp

	// timestap := strconv.FormatInt(time.Now().UnixNano(), 10)
	// counter++
	filename := fmt.Sprintf("chunk-%d.webm", sessionCounters[sessionID])
	filepath := filepath.Join(basepath, "webm", filename)

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

	// ✅ Ensure we flush and close file properly
	err = dst.Sync() // flush buffered data to disk
	if err != nil {
		http.Error(w, "Failed to sync file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = dst.Close() // close file descriptor
	if err != nil {
		http.Error(w, "Failed to close file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Wait until the file is stable before conversion
	stable := waitForFileStability(filepath, 500*time.Millisecond)
	if !stable {
		http.Error(w, "File is not stable for conversion", http.StatusInternalServerError)
		return
	}

	convertQueue <- ConvertTask{
		FilePath:  filepath,
		SessionID: sessionID,
		ChunkID:   int(sessionCounters[sessionID]),
	}
	log.Printf("Saved chunk : %s (%s) \n", filename, handler.Header.Get("Content-Type"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chunk uploaded successfully"))

}

func convertChunkToHLS(chunkPath, sessionID string, chunkNumber int) error {
	//prepare paths
	log.Println("XXXXX________XXXXX_____", chunkPath, chunkNumber)
	basedir := filepath.Join("uploads", "session-"+sessionID, "hls")
	outputTSFile := fmt.Sprintf("chunk-%d", chunkNumber)
	outputTSPath := filepath.Join(basedir, outputTSFile+".ts")
	playlistPath := filepath.Join(basedir, "output.m3u8")
	log.Println("--------------------------------------", outputTSPath, playlistPath)
	cmd := exec.Command("ffmpeg",
		"-i", chunkPath,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-c:a", "aac",
		"-strict", "experimental",
		outputTSPath,
	)

	// Optional: print ffmpeg output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Converting %s to %s\n", chunkPath, outputTSPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	// Step 2: Append to playlist
	playlistEntry := fmt.Sprintf("#EXTINF:1.0,\n%s\n", outputTSFile)

	// Check if playlist exists, if not, create and write header
	if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
		header := "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:1\n#EXT-X-MEDIA-SEQUENCE:0\n"
		if err := os.WriteFile(playlistPath, []byte(header), 0644); err != nil {
			return fmt.Errorf("failed to create playlist: %w", err)
		}
	}

	// Append the new chunk to the playlist
	playlistFile, err := os.OpenFile(playlistPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open playlist: %w", err)
	}
	defer playlistFile.Close()

	if _, err := playlistFile.WriteString(playlistEntry); err != nil {
		return fmt.Errorf("failed to write to playlist: %w", err)
	}

	log.Printf("Appended %s to %s\n", outputTSFile, playlistPath)
	return nil
}
