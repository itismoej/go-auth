package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mjafari98/go-auth/models"
	"github.com/mjafari98/go-auth/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

const (
	GRPCPort = ":50051"
	RESTPort = ":9090"
)

var DB = models.ConnectAndMigrate()

func main() {
	authServer := AuthServer{}

	// start REST server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	_ = pb.RegisterAuthHandlerServer(ctx, mux, &authServer)

	log.Printf(
		"server REST started in localhost%s (Wait 60 second before making http requests) ...\n",
		RESTPort,
	)
	go func() {
		err := http.ListenAndServe(RESTPort, mux)
		if err != nil {
			log.Fatal("cannot start REST server: ", err)
		}
	}()
	// end of REST server

	// start gRPC server
	listener, err := net.Listen("tcp", GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServer(grpcServer, &authServer)
	reflection.Register(grpcServer)

	log.Printf("server gRPC is starting in localhost%s ...\n", GRPCPort)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start GRPC server: ", err)
	}
	// end of gRPC server
}
