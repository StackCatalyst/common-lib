package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/StackCatalyst/common-lib/examples/calculator"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CalculatorServer implements the Calculator service
type CalculatorServer struct {
	pb.UnimplementedCalculatorServer
}

// Add implements the Add RPC method
func (s *CalculatorServer) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	result := req.A + req.B
	return &pb.AddResponse{Result: result}, nil
}

// Subtract implements the Subtract RPC method
func (s *CalculatorServer) Subtract(ctx context.Context, req *pb.SubtractRequest) (*pb.SubtractResponse, error) {
	result := req.A - req.B
	return &pb.SubtractResponse{Result: result}, nil
}

func main() {
	// Example 1: Create and start gRPC server
	server := grpc.NewServer()
	pb.RegisterCalculatorServer(server, &CalculatorServer{})

	// Start the server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Example 2: Create gRPC client
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a Calculator client
	calculatorClient := pb.NewCalculatorClient(conn)

	// Example 3: Make gRPC calls with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test Add RPC
	addReply, err := calculatorClient.Add(ctx, &pb.AddRequest{A: 10, B: 5})
	if err != nil {
		log.Printf("Failed to call Add: %v", err)
	} else {
		fmt.Printf("Add result: %d\n", addReply.Result)
	}

	// Test Subtract RPC
	subReply, err := calculatorClient.Subtract(ctx, &pb.SubtractRequest{A: 10, B: 5})
	if err != nil {
		log.Printf("Failed to call Subtract: %v", err)
	} else {
		fmt.Printf("Subtract result: %d\n", subReply.Result)
	}

	// Example 4: Graceful server shutdown
	server.GracefulStop()
}
