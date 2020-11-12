package main

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mjafari98/go-auth/models"
	"github.com/mjafari98/go-auth/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"io/ioutil"
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

func loadKey(pemData []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(pemData)

	if block == nil {
		return nil, fmt.Errorf("unable to load key")
	}

	if block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("wrong type of key - %s", block.Type)
	}

	return x509.ParseECPrivateKey(block.Bytes)
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

	data, err := ioutil.ReadFile("ecdsa-p521-private.pem")
	if err != nil {
		panic(err)
	}
	privateKey, err := loadKey(data)
	if err != nil {
		panic(err)
	}

	return token.SignedString(privateKey)
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

var DB = models.ConnectAndMigrate()
var jwtManager = JWTManager{
	secretKey:     secretKey,
	tokenDuration: tokenDuration,
}

type AuthServer struct {
	pb.UnimplementedAuthServer
}

func (server *AuthServer) Login(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	var user models.User
	result := DB.Take(&user, "username = ?", credentials.GetUsername())
	if errors.Is(result.Error, gorm.ErrRecordNotFound) || !user.PasswordIsCorrect(credentials.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := jwtManager.Generate(&user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.Token{Access: token}
	return res, nil
}

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServer(grpcServer, &AuthServer{})
	reflection.Register(grpcServer)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
