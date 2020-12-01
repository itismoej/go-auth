package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
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
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 24 * time.Hour
	port                 = ":50051"
)

type JWTManager struct {
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

func (manager *JWTManager) Generate(user *models.User) *pb.JWTToken {
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

	key, err := ioutil.ReadFile("cs/server/ecdsa-p521-private.pem")
	if err != nil {
		panic(err)
	}
	privateKey, err := jwt.ParseECPrivateKeyFromPEM(key)
	if err != nil {
		panic(err)
	}

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	return &pb.JWTToken{Token: signedToken}
}

func (manager *JWTManager) Verify(jwtToken string) (*UserClaims, error) {
	var err error

	publicKeyPath := "cs/server/ecdsa-p521-public.pem"
	key, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse ECDSA public key: %v", publicKeyPath)
	}

	var ecdsaKey *ecdsa.PublicKey
	if ecdsaKey, err = jwt.ParseECPublicKeyFromPEM(key); err != nil {
		return nil, fmt.Errorf("unable to parse ECDSA public key: %v", err)
	}

	parts := strings.Split(jwtToken, ".")

	err = jwt.SigningMethodES512.Verify(strings.Join(parts[0:2], "."), parts[2], ecdsaKey)
	if err != nil {
		return nil, fmt.Errorf("error while verifying key: %v", err)
	}

	token, err := jwt.ParseWithClaims(jwtToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return ecdsaKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid claims: %v", ok)
	}
}

var DB = models.ConnectAndMigrate()

var accessJwtManager = JWTManager{
	tokenDuration: accessTokenDuration,
}
var refreshJwtManager = JWTManager{
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
	claims, err := refreshJwtManager.Verify(refreshToken.Token)
	if err != nil {
		fmt.Println(err)
		return nil, status.Errorf(codes.Aborted, "jwt is not valid")
	}

	var user models.User
	result := DB.Take(&user, "username = ?", claims.Username)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.NotFound, "incorrect claims")
	}

	access := accessJwtManager.Generate(&user)

	res := &pb.JWTToken{Token: access.Token}
	return res, nil
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
