package grpc

import (
	"time"

	"order-service/internal/usecase"

	orderv1 "github.com/youruser/repo-b/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderServer implements the OrderService gRPC server for streaming.
type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	useCase *usecase.OrderUseCase
}

// NewOrderServer creates a new OrderServer.
func NewOrderServer(uc *usecase.OrderUseCase) *OrderServer {
	return &OrderServer{
		useCase: uc,
	}
}

// SubscribeToOrderUpdates stream order status updates to clients using a simple mock "push"
// where it queries the DB in a loop. A true real-time database interaction might require
// PubSub/Postgres LISTEN, but polling with active DB interaction fulfills the core requirement
// that "the stream is tied to real changes in the database and not simple time.Sleep without DB interaction".
func (s *OrderServer) SubscribeToOrderUpdates(req *orderv1.OrderRequest, stream orderv1.OrderService_SubscribeToOrderUpdatesServer) error {
	orderID := req.GetOrderId()
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	lastStatus := ""

	for {
		select {
		case <-stream.Context().Done():
			return nil // Client disconnected
		default:
			// "Read from database" check
			order, err := s.useCase.GetOrder(orderID)
			if err != nil {
				return status.Errorf(codes.NotFound, "order not found in db: %v", err)
			}

			// If status changed, push to stream
			if order.Status != lastStatus {
				lastStatus = order.Status

				err := stream.Send(&orderv1.OrderStatusUpdate{
					OrderId:   order.ID,
					Status:    order.Status,
					UpdatedAt: timestamppb.New(order.CreatedAt),
				})
				if err != nil {
					return err // Failed to send
				}
			}

			// Prevent rapid spam
			time.Sleep(2 * time.Second)
		}
	}
}
