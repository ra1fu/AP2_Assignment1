package grpc

import (
	"context"

	"payment-service/internal/usecase"

	paymentv1 "github.com/ra1fu/ap2-repo-b/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentServer implements the PaymentService gRPC server.
type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

// NewPaymentServer creates a new gRPC Payment Server.
func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	return &PaymentServer{
		useCase: uc,
	}
}

// ProcessPayment handles the incoming gRPC request to authorize a payment.
func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than zero")
	}

	// In the usecase, amount is int64 (usually cents).
	// Our proto has double. We'll cast to int64 based on your domain model logic.
	amountInCents := int64(req.GetAmount() * 100)

	payment, err := s.useCase.AuthorizePayment(req.GetOrderId(), amountInCents)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process payment: %v", err)
	}

	return &paymentv1.PaymentResponse{
		PaymentId:   payment.ID,
		Status:      payment.Status,
		ProcessedAt: timestamppb.New(payment.CreatedAt),
	}, nil
}

// ListPayments handles the incoming gRPC request to fetch a list of payments by amount range.
func (s *PaymentServer) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	if req.GetMinAmount() > 0 && req.GetMaxAmount() > 0 && req.GetMinAmount() > req.GetMaxAmount() {
		return nil, status.Error(codes.InvalidArgument, "min_amount cannot be greater than max_amount")
	}

	payments, err := s.useCase.ListPayments(req.GetMinAmount(), req.GetMaxAmount())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list payments: %v", err)
	}

	var pbPayments []*paymentv1.PaymentResponse
	for _, payment := range payments {
		pbPayments = append(pbPayments, &paymentv1.PaymentResponse{
			PaymentId:   payment.ID,
			Status:      payment.Status,
			ProcessedAt: timestamppb.New(payment.CreatedAt),
		})
	}

	return &paymentv1.ListPaymentsResponse{
		Payments: pbPayments,
	}, nil
}
