package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (s *Server) handleFileTransfer(conn net.Conn, addr string) {
	defer conn.Close()

	go s.readLoop(conn)
	s.writeLoop(conn)
}

func (s *Server) readLoop(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msgType, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}
		msgType = strings.TrimSpace(msgType)

		switch msgType {
		case "CHAT":
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}
			fmt.Println("Received chat message:", strings.TrimSpace(message))
		case "FILE":
			if err := s.handleFileTransfer(conn, reader); err != nil {
				fmt.Println("File transfer error:", err)
				return
			}
		default:
			fmt.Println("Unknown message type:", msgType)
		}
	}
}

func (s *Server) handleFileTransfer(conn net.Conn, reader *bufio.Reader) error {
	// Read the file name size
	fileNameSizeStr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	fileNameSize, err := strconv.ParseInt(strings.TrimSpace(fileNameSizeStr), 10, 64)
	if err != nil {
		return err
	}

	// Read the file name
	fileName := make([]byte, fileNameSize)
	_, err = io.ReadFull(reader, fileName)
	if err != nil {
		return err
	}

	// Read the file size
	fileSizeStr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	fileSize, err := strconv.ParseInt(strings.TrimSpace(fileSizeStr), 10, 64)
	if err != nil {
		return err
	}

	// Ensure the save directory exists
	err = os.MkdirAll(s.saveDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Create a new file to save the received data
	outFile, err := os.Create(filepath.Join(s.saveDir, string(fileName)))
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Read the file data
	_, err = io.CopyN(outFile, reader, fileSize)
	if err != nil {
		return err
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
