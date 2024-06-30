package main

import (
	"fmt"
	"ws/server"
)

func main() {
	s := server.NewServer(":3000", "./saved_files")

	go s.Start()

	fmt.Println("Server started successfully")

	for {

	}
}
