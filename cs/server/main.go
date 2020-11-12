package main

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mjafari98/go-auth/models"
	"github.com/mjafari98/go-auth/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
	"log"
	"net"
	"time"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
	port          = ":50051"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	Username  string      `json:"username"`
	Role      models.Role `json:"role"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{secretKey, tokenDuration}
}

func (manager *JWTManager) Generate(user *models.User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Username:  user.Username,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	return token.SignedString([]byte(manager.secretKey))
}

func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodECDSA)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

type AuthServer struct {
	pb.UnimplementedAuthServer
	//db         *gorm.DB
	//jwtManager *JWTManager
}

func NewAuthServer(db *gorm.DB, jwtManager *JWTManager) *AuthServer {
	//return &AuthServer{db, jwtManager}
	return &AuthServer{}
}

func (server *AuthServer) Login(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	//var user models.User
	//result := server.db.Take(&user, "username = ?", credentials.GetUsername())
	//if errors.Is(result.Error, gorm.ErrRecordNotFound) || !user.CheckPassword(credentials.GetPassword()) {
	//	return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	//}
	//
	//token, err := server.jwtManager.Generate(&user)
	//if err != nil {
	//	return nil, status.Errorf(codes.Internal, "cannot generate access token")
	//}
	//
	//res := &pb.Token{Access: token}
	res := &pb.Token{Access: "salam"}
	return res, nil
}

func main() {
	db := models.ConnectAndMigrate()
	jwtManager := NewJWTManager(secretKey, tokenDuration)
	authServer := NewAuthServer(db, jwtManager)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
