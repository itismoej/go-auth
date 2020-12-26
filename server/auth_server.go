package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mjafari98/go-auth/models"
	"github.com/mjafari98/go-auth/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"time"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 24 * time.Hour
)

var accessJwtManager = NewJWTManager(accessTokenDuration)
var refreshJwtManager = NewJWTManager(refreshTokenDuration)

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
	creator := ctx.Value("user")
	if creator == nil || !creator.(models.User).IsAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied: Only Admin can create user")
	}

	var newUser models.User
	newUser.FillFromProtoBuf(user)
	newUser.IsActive = true
	newUser.SetNewPassword(user.Password)

	result := DB.Create(&newUser)
	if errors.Is(result.Error, gorm.ErrInvalidData) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid data has been entered")
	} else if errors.Is(result.Error, gorm.ErrRegistered) {
		return nil, status.Errorf(codes.AlreadyExists, "this user is already registered")
	} else if result.Error != nil {
		return nil, status.Errorf(codes.Unknown, "%s", result.Error)
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
