//go:build !js

package game

import (
	"fmt"
	ipc "github.com/james-barrow/golang-ipc"
)

type TextServer struct {
	server      *ipc.Server
	socket      string
	messageType int
}

func NewTextServer(socket string, msgType int) *TextServer {
	server, _ := ipc.StartServer(socket, nil)
	return &TextServer{
		socket:      socket,
		messageType: msgType,
		server:      server,
	}
}

func (s *TextServer) Write(text string) {
	if s.server == nil || (s.server.StatusCode() != ipc.Connected && s.server.StatusCode() != ipc.Listening) {
		s.server, _ = ipc.StartServer(s.socket, nil)
	}
	if s.server.StatusCode() == ipc.Connected {
		err := s.server.Write(s.messageType, []byte(text))
		if err != nil {
			println(fmt.Sprintf("error writing to log server: %v", err))
			return
		}
	}
}
