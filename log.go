package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// SizeLimitedLogWriter is a custom writer that ensures a log file remains within a specified size limit.
type SizeLimitedLogWriter struct {
	filePath   string     // Path to the log file
	maxSize    int64      // Maximum file size in bytes
	keepBytes  int64      // Number of recent bytes to retain when truncating
	currentLog *os.File   // The current log file
	mu         sync.Mutex // Mutex to ensure thread-safe operations
}

// NewSizeLimitedLogWriter creates a new instance of SizeLimitedLogWriter.
// filePath: the log file path
// maxSizeMB: the maximum allowed size of the log file in megabytes
// keepSizeMB: the amount of recent data to retain in megabytes
func NewSizeLimitedLogWriter(filePath string, maxSizeMB, keepSizeMB int) (*SizeLimitedLogWriter, error) {
	// Convert sizes from megabytes to bytes
	maxSize := int64(maxSizeMB) * 1024 * 1024
	keepBytes := int64(keepSizeMB) * 1024 * 1024

	// Validate that keepBytes does not exceed maxSize
	if keepBytes > maxSize {
		return nil, fmt.Errorf("keepSizeMB can't be greather than maxSizeMB") // Returning error if it doesn't make sense to keep more than max size
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}

	// Open or create the initial log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &SizeLimitedLogWriter{
		filePath:   filePath,
		maxSize:    maxSize,
		keepBytes:  keepBytes,
		currentLog: file,
	}, nil
}

// Write writes data to the log file and enforces size limits.
func (w *SizeLimitedLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock() // Lock the writer to prevent concurrent writes
	defer w.mu.Unlock()

	// Check the current size of the log file
	info, err := w.currentLog.Stat()
	if err != nil {
		return 0, err
	}

	// If the file exceeds the maximum size, truncate it
	if info.Size() > w.maxSize {
		// Close the current log file
		w.currentLog.Close()

		// Retain only the most recent logs and truncate the file
		if err := w.truncateAndKeepRecentLogs(); err != nil {
			return 0, err
		}

		// Reopen the truncated log file
		w.currentLog, err = os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
	}

	// Write the new log data to the file
	return w.currentLog.Write(p)
}

// truncateAndKeepRecentLogs retains the most recent logs in the file while truncating older data.
func (w *SizeLimitedLogWriter) truncateAndKeepRecentLogs() error {
	// Open the file for reading
	file, err := os.Open(w.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a buffer to store the most recent `keepBytes` data
	var buffer bytes.Buffer
	_, err = file.Seek(-w.keepBytes, io.SeekEnd) // Move the file pointer to the last `keepBytes` bytes
	if err != nil {
		return err
	}

	_, err = io.Copy(&buffer, file) // Read the data into the buffer
	if err != nil {
		return err
	}

	// Truncate the file to remove older logs
	if err := os.Truncate(w.filePath, 0); err != nil {
		return err
	}

	// Reopen the file in write mode and write back the retained logs
	file, err = os.OpenFile(w.filePath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes()) // Write the recent logs back into the file
	return err
}

// Close closes the log file safely.
func (w *SizeLimitedLogWriter) Close() error {
	w.mu.Lock() // Lock to ensure no operations are ongoing during closure
	defer w.mu.Unlock()
	return w.currentLog.Close()
}
