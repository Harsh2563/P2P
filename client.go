package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	serverAddr := "192.168.64.251:3000" // Replace with the server's IP address
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	go readLoop(conn)
	writeLoop(conn)
}

func readLoop(conn net.Conn) {
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
			if err := handleFileTransfer(conn, reader); err != nil {
				fmt.Println("File transfer error:", err)
				return
			}
		default:
			fmt.Println("Unknown message type:", msgType)
		}
	}
}

func handleFileTransfer(conn net.Conn, reader *bufio.Reader) error {
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

	// Create a new file to save the received data
	outFile, err := os.Create(string(fileName))
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

func writeLoop(conn net.Conn) {
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
