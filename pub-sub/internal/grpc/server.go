package grpc

import (
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"google.golang.org/grpc"
)

type serverAPI struct {
	pb.UnimplementedPubSubServiceServer
	service types.PubSubServiceInterface
}

func Register(srv *grpc.Server, service types.PubSubServiceInterface) {
	pb.RegisterPubSubServiceServer(srv, &serverAPI{service: service})
}

const (
	emptyValue = 0
)

func (h *serverAPI) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	return h.service.Subscribe(req, stream)
}

func validateChat(chatID string) {
	if chatID == "" {

	}
}
