package service

import (
	"context"
	"errors"
	"github.com/mjafari98/go-auth/models"
	pb "github.com/mjafari98/go-auth/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthServer struct {
	db         *gorm.DB
	jwtManager *JWTManager
}

func NewAuthServer(db *gorm.DB, jwtManager *JWTManager) *AuthServer {
	return &AuthServer{db, jwtManager}
}

func (server *AuthServer) Login(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	var user models.User
	result := server.db.Take(&user, "username = ?", credentials.GetUsername())
	if errors.Is(result.Error, gorm.ErrRecordNotFound) || !user.CheckPassword(credentials.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := server.jwtManager.Generate(&user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.Token{Access: token}
	return res, nil
}
