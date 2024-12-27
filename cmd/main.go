package main

import (
	"go-grpc/cmd/config"
	"go-grpc/cmd/services"
	productPb "go-grpc/pb/product"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	// init listener
	netListen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}

	// load .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// init database connection
	db := config.ConnectDatabase()

	// init gRPC server
	grpcServer := grpc.NewServer()
	productService := services.ProductService{DB: db}
	productPb.RegisterProductServiceServer(grpcServer, &productService)

	log.Printf("Server listening at %v", netListen.Addr())

	// serve gRPC server
	if err := grpcServer.Serve(netListen); err != nil {
		log.Fatalf("failed to serve: %v", err.Error())
	}
}
