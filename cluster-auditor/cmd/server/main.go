package main

import (
	"log"
	"net"

	"goprojects/services/generated/auditorpb"
	"goprojects/services/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	db, err := server.InitDB("audit.db")
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	srv := &server.AuditorServer{DB: db}

	grpcServer := grpc.NewServer()
	auditorpb.RegisterClusterAuditorServer(grpcServer, srv)

	reflection.Register(grpcServer)

	log.Println("gRPC server listening on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
