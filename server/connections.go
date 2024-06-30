package server

import (
	"fmt"
)

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		addr := conn.RemoteAddr().String()
		fmt.Println("Accept connection:", addr)

		s.saveUser(addr, true)
		fmt.Println("Users: ", s.users)

		go s.handleFileTransfer(conn, addr)
	}
}

func (s *Server) saveUser(user string, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user] = online
}
