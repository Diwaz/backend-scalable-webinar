package main

type ConvertTask struct {
	FilePath  string
	SessionID string
	ChunkID   int
}

var convertQueue = make(chan ConvertTask, 100)

func init() {
	go worker()
}

func worker() {
	for task := range convertQueue {
		convertChunkToHLS(task.FilePath, task.SessionID, task.ChunkID)

		// convertChunkToHLS(filepath, sessionID, int(sessionCounters[sessionID]))
	}
}
