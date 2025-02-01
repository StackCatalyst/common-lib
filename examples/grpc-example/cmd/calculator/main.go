package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/StackCatalyst/common-lib/examples/grpc-example/internal/calculator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// CalculatorServer implements the Calculator service
type CalculatorServer struct {
	pb.UnimplementedCalculatorServer
}

// Add implements the Add RPC method
func (s *CalculatorServer) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	// Example of context deadline handling
	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.DeadlineExceeded, "request timed out")
	}

	// Example of input validation
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Example of business logic validation
	if req.A > 1000000 || req.B > 1000000 {
		return nil, status.Error(codes.InvalidArgument, "numbers too large, maximum allowed is 1,000,000")
	}

	result := req.A + req.B
	return &pb.AddResponse{Result: result}, nil
}

// Subtract implements the Subtract RPC method
func (s *CalculatorServer) Subtract(ctx context.Context, req *pb.SubtractRequest) (*pb.SubtractResponse, error) {
	// Example of context deadline handling
	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.DeadlineExceeded, "request timed out")
	}

	// Example of input validation
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Example of business logic validation
	if req.A > 1000000 || req.B > 1000000 {
		return nil, status.Error(codes.InvalidArgument, "numbers too large, maximum allowed is 1,000,000")
	}

	result := req.A - req.B
	return &pb.SubtractResponse{Result: result}, nil
}

func main() {
	// Example 1: Create and start gRPC server with interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor()),
	)
	pb.RegisterCalculatorServer(server, &CalculatorServer{})

	// Start the server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		log.Printf("Server listening on :50051")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Example 2: Create gRPC client with timeout and retry options
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Wait for connection to be established
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a Calculator client
	calculatorClient := pb.NewCalculatorClient(conn)

	// Example 3: Make gRPC calls with proper error handling
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test Add RPC with valid input
	addReply, err := calculatorClient.Add(ctx, &pb.AddRequest{A: 10, B: 5})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("Failed to call Add: code=%v, message=%v", st.Code(), st.Message())
		} else {
			log.Printf("Failed to call Add: %v", err)
		}
	} else {
		fmt.Printf("Add result: %d\n", addReply.Result)
	}

	// Test Add RPC with invalid input
	addReply, err = calculatorClient.Add(ctx, &pb.AddRequest{A: 2000000, B: 5})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("Expected error from Add: code=%v, message=%v", st.Code(), st.Message())
		} else {
			log.Printf("Failed to call Add: %v", err)
		}
	}

	// Test Subtract RPC
	subReply, err := calculatorClient.Subtract(ctx, &pb.SubtractRequest{A: 10, B: 5})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("Failed to call Subtract: code=%v, message=%v", st.Code(), st.Message())
		} else {
			log.Printf("Failed to call Subtract: %v", err)
		}
	} else {
		fmt.Printf("Subtract result: %d\n", subReply.Result)
	}

	// Example 4: Graceful server shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Printf("Server shutdown timed out, forcing stop")
		server.Stop()
	case <-done:
		log.Printf("Server shutdown gracefully")
	}
}

// unaryInterceptor returns a server interceptor function to handle logging and metrics
func unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		method := info.FullMethod

		log.Printf("Request - Method: %s", method)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("Response - Method: %s, Duration: %v, Error: code=%v message=%v",
				method, duration, st.Code(), st.Message())
		} else {
			log.Printf("Response - Method: %s, Duration: %v", method, duration)
		}

		return resp, err
	}
}
