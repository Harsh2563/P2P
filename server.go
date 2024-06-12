package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Message struct {
        from string
        data []byte
}

type Server struct {
	listenAddr string
	listner    net.Listener
        quitchan   chan struct{}
        msgchan    chan Message
}

func NewServer (listenAddr string) *Server {
        return &Server {
                listenAddr: listenAddr,
                quitchan: make(chan struct{}),
                msgchan: make(chan Message,10),
        }
}

func (s* Server) Start() error {
        ln,err := net.Listen("tcp", s.listenAddr)
        if err != nil {
                return err
        }
        
        s.listner = ln

        defer s.listner.Close()

        go s.acceptLoop()

        <-s.quitchan
        close(s.msgchan)

        return nil
}

func (s* Server) acceptLoop() {
        for {
                conn,err := s.listner.Accept()
                if err != nil {
                        fmt.Println("Accept error: ", err)
                        continue
                }  

                fmt.Println("Accept connection: ", conn.RemoteAddr().String())

                go s.readLoop(conn)
        }
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error: ", err)
			return 
		}
		fmt.Println("Read message: ", msg)
                s.msgchan <- Message{from: conn.RemoteAddr().String(), data: []byte(msg)}
	}
}

func main() {
        server := NewServer(":3000")

        go func() {
                for msg:= range server.msgchan {
                        fmt.Printf("Message from (%s): %s\n", msg.from,string(msg.data))
                }
        } ()
        log.Fatal(server.Start())
}