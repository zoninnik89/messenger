package websocketserver

import (
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
	// Access the cookies
	cookie, err := r.Cookie("token")
	if err != nil {
		s.logger.Errorw("missing or invalid JWT token cookie", "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	jwtToken := cookie.Value

	// remove the "Bearer " prefix if present
	if len(jwtToken) > 7 && jwtToken[:7] == "Bearer " {
		jwtToken = jwtToken[7:]
	}

	userID, err := s.validateJWTToken(jwtToken)
	if err != nil {
		s.logger.Errorw("invalid JWT token", "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket if the token is valid
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("failed to upgrade to WebSocket", "error", err)
		return
	}

	s.HandleWS(ws, r, userID)
}

func (s *WebsocketServer) HandleWS(ws *websocket.Conn, r *http.Request, userID string) {
	const op = "websocketserver.handleWS"
	s.logger.Infow("new incomming connection from client", "op", op, "addr", ws.RemoteAddr())

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
			if err := ws.WriteMessage(websocket.TextMessage, messageData); err != nil {
				s.logger.Errorw("failed to send message to WebSocket", "op", op, "error", err)
				break
			}
		}
	}()

	s.ReadLoop(ws, userID)

}

func (s *WebsocketServer) ReadLoop(ws *websocket.Conn, userID string) {
	const op = "websocketserver.readLoop"

	for {
		// Read a message from the WebSocket
		messageType, messageData, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Errorw("unexpected WebSocket closure", "op", op, "err", err)
			} else {
				s.logger.Infow("client closed WebSocket connection", "op", op, "err", err)
			}
			break
		}
		s.logger.Infow("message received", "op", op, "userID", userID, "messageType", messageType, "message", string(messageData))

		var messageParsed Message
		err = json.Unmarshal(messageData, &messageParsed)
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
