package main

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mjafari98/go-auth/models"
	"github.com/mjafari98/go-auth/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	secretKey     = "secret"
	accessTokenDuration = 15 * time.Minute
	refreshTokenDuration = 24 * time.Hour
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

type UserClaimsRefresh struct {
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

func (manager *JWTManager) Generate(user *models.User) (*pb.JWTToken) {
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

	data, err := ioutil.ReadFile("cs/server/ecdsa-p521-private.pem")
	if err != nil {
		panic(err)
	}
	privateKey, err := loadKey(data)
	if err != nil {
		panic(err)
	}

	signedToken, err := token.SignedString(privateKey) 
	if err != nil {
		panic(err)
	}

	return &pb.JWTToken{Token: signedToken}
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

var accessJwtManager = JWTManager{
	secretKey:     secretKey,
	tokenDuration: accessTokenDuration,
}
var refreshJwtManager = JWTManager{
	secretKey:     secretKey,
	tokenDuration: refreshTokenDuration,
}


type AuthServer struct {
	pb.UnimplementedAuthServer
}

func (server *AuthServer) Login(ctx context.Context, credentials *pb.Credentials) (*pb.PairToken, error) {
	var user models.User
	result := DB.Take(&user, "username = ?", credentials.GetUsername())
	if errors.Is(result.Error, gorm.ErrRecordNotFound) || !user.PasswordIsCorrect(credentials.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	accessToken := accessJwtManager.Generate(&user)
	refreshToken := refreshJwtManager.Generate(&user)

	res := &pb.PairToken{Access: accessToken, Refresh: refreshToken}
	return res, nil
}

func (server *AuthServer) Signup(ctx context.Context, user *pb.User) (*pb.User, error) {
	var newUser models.User
	newUser.FromProtoBuf(user)
	newUser.IsActive = true
	newUser.SetNewPassword(user.Password)

	result := DB.Create(&newUser)
	if errors.Is(result.Error, gorm.ErrInvalidData) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid data has been entered")
	}
	if errors.Is(result.Error, gorm.ErrRegistered) {
		return nil, status.Errorf(codes.AlreadyExists, "this user is already registered")
	}

	user = newUser.ConvertToProtoBuf()
	return user, nil
}

func (server *AuthServer) RefreshAccessToken(ctx context.Context, refreshToken *pb.JWTToken) (*pb.JWTToken, error) {

}

func main() {
	authServer := AuthServer{}

	// start REST server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	_ = pb.RegisterAuthHandlerServer(ctx, mux, &authServer)

	log.Println("server REST started in localhost:9090 (Wait 60 second before making http requests) ...")
	go http.ListenAndServe(":9090", mux)
	// end of REST server

	// start gRPC server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServer(grpcServer, &authServer)
	reflection.Register(grpcServer)

	log.Println("server gRPC is starting in localhost:50051 ...")
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
	// end of gRPC server
}
