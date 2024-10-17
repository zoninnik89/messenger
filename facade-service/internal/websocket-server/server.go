package websocketserver

import (
	"io"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	pb "github.com/zoninnik89/messenger/common/api"
	grpcgateway "github.com/zoninnik89/messenger/facade-service/internal/gateway"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

type WebsocketServer struct {
	logger *zap.SugaredLogger
	gw     *grpcgateway.Gateway
	conns  map[*websocket.Conn]bool
}

func NewWebsocketServer(g *grpcgateway.Gateway) *WebsocketServer {
	l := logging.GetLogger().Sugar()
	return &WebsocketServer{
		conns:  make(map[*websocket.Conn]bool),
		gw:     g,
		logger: l,
	}
}

type Message struct {
	ChatID      string `json:"chat_id"`
	MessageText string `json:"message_text"`
}

func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("failed to upgrade to WebSocket", "error", err)
		return
	}

	// Pass the original HTTP request to the HandleWS function
	s.HandleWS(ws, r)
}

func (s *WebsocketServer) HandleWS(ws *websocket.Conn, r *http.Request) {
	const op = "websocketserver.handleWS"
	s.logger.Infow("new incomming connection from client", "op", op, "addr", ws.RemoteAddr())

	// Get the JWT token from the request header
	jwtToken := r.Header.Get("Authorization")
	if jwtToken == "" {
		s.logger.Errorw("missing Authorization header", "op", op)
		ws.Close()
		return
	}

	// remove the "Bearer " prefix if present
	if len(jwtToken) > 7 && jwtToken[:7] == "Bearer " {
		jwtToken = jwtToken[7:]
	}

	userID, err := s.validateJWTToken(jwtToken)
	if err != nil {
		s.logger.Errorw("invalid JWT token", "op", op, "error", err)
		ws.Close()
		return
	}

	s.conns[ws] = true // add mutex

	messagesChan := make(chan *pb.Message)

	// Start a goroutine to establish the gRPC stream and read messages
	go func() {
		err := s.gw.GetMessagesStream(context.Background(), &pb.GetMessagesStreamRequest{UserId: userID}, messagesChan)
		if err != nil {
			s.logger.Errorw("failed to get message stream", "op", op, "error", err)
			ws.Close()
			return
		}
	}()

	// Start a goroutine to read messages from the gRPC stream and send to WebSocket
	go func() {
		for msg := range messagesChan {
			// Serialize the message to JSON
			messageData, err := json.Marshal(msg)
			if err != nil {
				s.logger.Errorw("failed to marshal message", "op", op, "error", err)
				continue
			}

			// Send the message to the WebSocket
			if err := ws.Write(websocket.TextMessage, messageData); err != nil {
				s.logger.Errorw("failed to send message to WebSocket", "op", op, "error", err)
				break
			}
		}
	}()

	s.ReadLoop(ws, userID)

}

func (s *WebsocketServer) ReadLoop(ws *websocket.Conn, userID string) {
	const op = "websocketserver.readLoop"

	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				s.logger.Infow("client closed WebSocket connection", "op", op, "err", err)
				break
			}
			s.logger.Errorw("error reading from WebSocket", "op", op, "err", err)
			continue
		}
		msg := buf[:n]
		s.logger.Infow("message received", "op", op, "message", string(msg))

		var messageParsed Message
		err = json.Unmarshal(msg, &messageParsed)
		if err != nil {
			s.logger.Errorw("error unmarshaling JSON message", "op", op, "err", err)
			continue
		}

		messageToBeSent := &pb.Message{
			ChatId:      messageParsed.ChatID,
			SenderId:    userID,
			MessageId:   uuid.NewString(),
			MessageText: messageParsed.MessageText,
			SentTs:      strconv.FormatInt(time.Now().Unix(), 10),
		}

		s.gw.SendMessage(
			context.Background(),
			&pb.SendMessageRequest{
				Message: messageToBeSent,
			},
		)

		ws.Write([]byte("message sent"))
	}

	s.cleanupConnection(ws)
}

func (s *WebsocketServer) cleanupConnection(ws *websocket.Conn) {
	// Remove the connection from the map and close the WebSocket
	delete(s.conns, ws)
	ws.Close()
}

func (s *WebsocketServer) validateJWTToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method and provide the secret key
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil // TO DO: Replace with a real secret going forward
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["uid"].(string)
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}
