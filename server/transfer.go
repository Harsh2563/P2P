package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

func (s *Server) handleFileTransfer(conn net.Conn, addr string) {
	defer conn.Close()
	defer s.saveUser(addr, false)
	// Read the file name size
	var fileNameSize int64
	err := readInt64(conn, &fileNameSize)
	if err != nil {
		fmt.Println("Failed to read file name size:", err)
		return
	}

	// Read the file name
	fileName := make([]byte, fileNameSize)
	_, err = io.ReadFull(conn, fileName)
	if err != nil {
		fmt.Println("Failed to read file name:", err)
		return
	}

	// Read the file size
	var fileSize int64
	err = readInt64(conn, &fileSize)
	if err != nil {
		fmt.Println("Failed to read file size:", err)
		return
	}

	// Ensure the save directory exists
	err = os.MkdirAll(s.saveDir, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create save directory:", err)
		return
	}

	// Create a new file to save the received data
	outFile, err := os.Create(filepath.Join(s.saveDir, string(fileName)))
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer outFile.Close()

	// Read the file data
	_, err = io.CopyN(outFile, conn, fileSize)
	if err != nil {
		fmt.Println("Failed to read file data:", err)
		return
	}

	fmt.Println("File received successfully:", string(fileName))

	// Save the transfer history
	s.history.AddEntry(addr, string(fileName), fileSize, time.Now())
}

func readInt64(conn net.Conn, n *int64) error {
	buf := make([]byte, 8)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return err
	}
	*n = int64(buf[0]) | int64(buf[1])<<8 | int64(buf[2])<<16 | int64(buf[3])<<24 | int64(buf[4])<<32 | int64(buf[5])<<40 | int64(buf[6])<<48 | int64(buf[7])<<56

	return nil
}
