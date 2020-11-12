package main

import (
	"github.com/mjafari98/go-auth/models"
	"github.com/mjafari98/go-auth/pb"
	"github.com/mjafari98/go-auth/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
	port = ":50051"

)

func main() {
	db := models.ConnectAndMigrate()
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(db, jwtManager)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
