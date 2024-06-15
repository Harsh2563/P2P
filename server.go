package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

type Server struct {
	listenAddr string
	listener   net.Listener
	quitchan   chan struct{}
	saveDir    string
}

func NewServer(listenAddr, saveDir string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitchan:   make(chan struct{}),
		saveDir:    saveDir,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	s.listener = ln

	defer s.listener.Close()

	go s.acceptLoop()

	<-s.quitchan

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		fmt.Println("Accept connection:", conn.RemoteAddr().String())

		go s.handleFileTransfer(conn)
	}
}

func (s *Server) handleFileTransfer(conn net.Conn) {
	defer conn.Close()

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

func main() {
	server := NewServer(":3000", "received_files")
	log.Fatal(server.Start())
}
