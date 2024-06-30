package server

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type TransferEntry struct {
	Addr      string
	FileName  string
	FileSize  int64
	Timestamp time.Time
}

type TransferHistory struct {
	entries []TransferEntry
	mu      sync.Mutex
}

func NewTransferHistory() *TransferHistory {
	return &TransferHistory{
		entries: make([]TransferEntry, 0),
	}
}

func (h *TransferHistory) AddEntry(addr, fileName string, fileSize int64, timestamp time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, TransferEntry{addr, fileName, fileSize, timestamp})
	h.saveToFile("transfer_history.txt")
}

func (h *TransferHistory) saveToFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open history file:", err)
		return
	}
	defer file.Close()

	for _, entry := range h.entries {
		line := fmt.Sprintf("%s\t%s\t%d\t%s\n", entry.Timestamp.Format(time.RFC3339), entry.Addr, entry.FileSize, entry.FileName)
		_, err = file.WriteString(line)
		if err != nil {
			fmt.Println("Failed to write history entry:", err)
		}
	}
}
