package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "main/proto"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	listeningAddress string
	artificialDelay  time.Duration
)

func init() {
	serverCmd.Flags().StringVar(&listeningAddress, "address", ":9090", "address to listen for gRPC connections on")
	serverCmd.Flags().DurationVar(&artificialDelay, "delay", 5*time.Millisecond, "artificial delay to add to each request to simulate work being done")

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

type server struct{}

func (*server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	time.Sleep(artificialDelay)

	log.Println("processed req")

	return &pb.HelloResponse{
		Message: "hello",
	}, nil
}

func runServer() {
	lis, err := net.Listen("tcp", listeningAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	log.Println("gRPC server is listening on port 9090...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
