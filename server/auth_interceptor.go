package main

import (
	"context"
	"fmt"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcContextTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/mjafari98/go-auth/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AuthInterceptorFunc(ctx context.Context) (context.Context, error) {
	accessToken, err := grpcAuth.AuthFromMD(ctx, "bearer")
	newCtx := context.WithValue(ctx, "user", nil)
	if err != nil {
		return newCtx, nil
	}
	claims, err := accessJwtManager.Verify(accessToken)
	if err != nil {
		fmt.Println(err)
		return nil, status.Errorf(codes.Unauthenticated, "jwt is not valid")
	}

	grpcContextTags.Extract(ctx).Set("auth.sub", claims)

	var user models.User
	result := DB.Joins("Role").Take(&user, "username = ?", claims.Username)
	if result.Error == nil {
		user.IsAdmin = user.Role.Name == "Admin"
		newCtx = context.WithValue(ctx, "user", user)
	}
	return newCtx, nil
}
