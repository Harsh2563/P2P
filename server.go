package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
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
	return nil
}

func (s *Server) writeLoop(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter message (or type 'FILE <filepath>' to send a file): ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		if strings.HasPrefix(input, "FILE ") {
			filePath := strings.TrimSpace(strings.TrimPrefix(input, "FILE "))
			if err := sendFile(writer, filePath); err != nil {
				fmt.Println("Failed to send file:", err)
			}
		} else {
			if _, err := writer.WriteString("CHAT\n" + input + "\n"); err != nil {
				fmt.Println("Failed to send chat message:", err)
				continue
			}
			writer.Flush()
		}
	}
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

func sendFile(writer *bufio.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	if _, err := writer.WriteString("FILE\n"); err != nil {
		return err
	}
	if _, err := writer.WriteString(fmt.Sprintf("%d\n", len(fileName))); err != nil {
		return err
	}
	if _, err := writer.WriteString(fileName); err != nil {
		return err
	}
	if _, err := writer.WriteString(fmt.Sprintf("%d\n", fileSize)); err != nil {
		return err
	}
	writer.Flush()

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}
	writer.Flush()

	fmt.Println("File sent successfully")
	return nil
}

func main() {
	server := NewServer(":3000", "received_files")
	log.Fatal(server.Start())
}
