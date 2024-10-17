package websocketserver

import (
	"io"

	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

type WebsocketServer struct {
	logger *zap.SugaredLogger
	conns  map[*websocket.Conn]bool
}

func NewWebsocketServer() *WebsocketServer {
	l := logging.GetLogger().Sugar()
	return &WebsocketServer{
		conns:  make(map[*websocket.Conn]bool),
		logger: l,
	}
}

func (s *WebsocketServer) HandleWS(ws *websocket.Conn) {
	const op = "websocketserver.handleWS"
	s.logger.Infow("new incomming connection from client", "op", op, "addr", ws.RemoteAddr())

	s.conns[ws] = true // add mutex
	s.ReadLoop(ws)

}

func (s *WebsocketServer) ReadLoop(ws *websocket.Conn) {
	const op = "websocketserver.readLoop"

	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			s.logger.Errorw("read error", "op", op, "err", err)
			continue
		}
		msg := buf[:n]
		s.logger.Infow("message received", "op", op, "message", msg)

		ws.Write([]byte("message sent"))
	}
}
