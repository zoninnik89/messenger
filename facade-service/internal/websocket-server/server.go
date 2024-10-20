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
	Type        string `json:"type"`
	ChatID      string `json:"chat_id"`
	MessageText string `json:"message_text"`
}

func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Access the cookies
	jwtToken := r.URL.Query().Get("token")
	if jwtToken == "" {
		s.logger.Errorw("missing or invalid JWT token", "error", "missing token", "token", jwtToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

	s.logger.Infow("valid JWT token received, proceeding with WebSocket upgrade", "userID", userID)

	// Upgrade to WebSocket if the token is valid
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("failed to upgrade to WebSocket", "error", err)
		return
	}

	s.logger.Infow("WebSocket upgrade successful", "userID", userID)

	s.HandleWS(ws, userID)
}

func (s *WebsocketServer) HandleWS(ws *websocket.Conn, userID string) {
	const op = "websocketserver.handleWS"

	ws.SetReadDeadline(time.Now().Add(60 * time.Second)) // Set the initial read deadline

	// Start a goroutine to send periodic ping messages
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pingMessage := map[string]string{
					"type": "ping",
				}
				messageData, _ := json.Marshal(pingMessage)
				if err := ws.WriteMessage(websocket.TextMessage, messageData); err != nil {
					s.logger.Errorw("failed to send ping message", "op", op, "error", err)
					ws.Close()
					return
				} else {
					//s.logger.Infow("ping message sent", "op", op)
				}
			}
		}
	}()

	s.conns[ws] = true // add mutex

	messagesChan := make(chan *pb.Message)
	done := make(chan struct{})

	errMessage, _ := json.Marshal("failed to get message stream on backend")

	// Start a goroutine to establish the gRPC stream and read messages
	go func() {
		err := s.gw.GetMessagesStream(context.Background(), &pb.GetMessagesStreamRequest{UserId: userID}, messagesChan)
		if err != nil {
			s.logger.Errorw("failed to get message stream", "op", op, "error", err)
			ws.WriteMessage(websocket.TextMessage, errMessage)
			ws.Close()
			return
		}
	}()

	// Start a goroutine to read messages from the gRPC stream and send to WebSocket
	go func() {
		for {
			select {
			case msg := <-messagesChan:
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
			case <-done:
				// Stop reading messages when done signal is received
				return
			}
		}
	}()

	// Run the ReadLoop in a separate goroutine
	go s.ReadLoop(ws, userID, done)

}

func (s *WebsocketServer) ReadLoop(ws *websocket.Conn, userID string, done chan struct{}) {
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
			errUnmarshalMessage, _ := json.Marshal(err.Error())
			ws.WriteMessage(websocket.TextMessage, errUnmarshalMessage)
			s.logger.Errorw("error unmarshaling JSON message", "op", op, "err", err)
			continue
		}

		if messageParsed.Type == "pong" {
			//s.logger.Infow("pong received from user", "op", op, "userID", userID)
			ws.SetReadDeadline(time.Now().Add(16 * time.Second)) // Extend the deadline when a pong is received
			continue
		}

		messageID := uuid.New().String()

		messageToBeSent := &pb.Message{
			ChatId:      messageParsed.ChatID,
			SenderId:    userID,
			MessageId:   messageID,
			MessageText: messageParsed.MessageText,
			SentTs:      strconv.FormatInt(time.Now().Unix(), 10),
		}

		_, err = s.gw.SendMessage(
			context.Background(),
			&pb.SendMessageRequest{
				Message: messageToBeSent,
			},
		)
		if err != nil {
			s.logger.Errorw("failed to send message to Chat client via GRPC", "op", op, "messageID", messageID, "error", err)
		} else {
			s.logger.Infow("successfully sent message to Chat client via GRPC", "op", op, "userID", userID, "messageID", messageID)
		}

	}

	close(done)
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
