package server

import (
	"net"
	"sync"
)

type Server struct {
	listenAddr string
	listener   net.Listener
	quitchan   chan struct{}
	saveDir    string
	users      map[string]bool
	mu         sync.Mutex
	history    *TransferHistory
}

func NewServer(listenAddr, saveDir string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitchan:   make(chan struct{}),
		saveDir:    saveDir,
		users:      make(map[string]bool),
		history:    NewTransferHistory(),
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
