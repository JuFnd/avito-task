package usecase

import (
	"avito-track/pkg/models"
	"avito-track/pkg/variables"
	"avito-track/services/authorization/proto/authorization"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
)

type IBannerRepository interface {
	GetBanners(featureID int64, tagIDs []int64, limit, offset int64) ([]models.Banner, error)
	UserBanner(tagID int64, featureID int64, useLastRevision bool) (*models.Banner, error)
}

type Core struct {
	logger            *slog.Logger
	bannersRepository IBannerRepository
	grpcClient        authorization.AuthorizationClient
}

func GetGrpcClient(address string) (authorization.AuthorizationClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf(variables.GrpcConnectError, ": %w", err)
	}
	client := authorization.NewAuthorizationClient(conn)

	return client, nil
}

func GetCore(configGrpc variables.GrpcConfig, banners IBannerRepository, logger *slog.Logger) *Core {
	client, err := GetGrpcClient(configGrpc.Address + ":" + configGrpc.Port)
	if err != nil {
		logger.Error(variables.GrpcConnectError, ": %w", err)
		return nil
	}
	return &Core{
		bannersRepository: banners,
		grpcClient:        client,
		logger:            logger,
	}
}

func (core *Core) UserBanner(tagID int64, featureID int64, useLastRevision bool) (*models.Banner, error) {
	banner, err := core.bannersRepository.UserBanner(tagID, featureID, useLastRevision)
	if err != nil {
		core.logger.Error(variables.BannerNotFoundError, ": %w", err)
		return nil, err
	}
	return banner, nil
}

func (core *Core) GetBanners(featureID int64, tagIDs []int64, limit, offset int64) ([]models.Banner, error) {
	banners, err := core.bannersRepository.GetBanners(featureID, tagIDs, limit, offset)
	if err != nil {
		core.logger.Error(variables.BannerNotFoundError, ": %w", err)
		return nil, err
	}

	return banners, nil
}

func (core *Core) GetUserRole(ctx context.Context, id int64) (string, error) {
	grpcRequest := authorization.RoleRequest{Id: id}

	grpcResponse, err := core.grpcClient.GetRole(ctx, &grpcRequest)
	if err != nil {
		core.logger.Error(variables.GrpcRecievError, err)
		return "", fmt.Errorf(variables.GrpcRecievError, err)
	}
	return grpcResponse.GetRole(), nil
}

func (core *Core) GetUserId(ctx context.Context, sid string) (int64, error) {
	grpcRequest := authorization.FindIdRequest{Sid: sid}

	grpcResponse, err := core.grpcClient.GetId(ctx, &grpcRequest)
	if err != nil {
		core.logger.Error(variables.GrpcRecievError, err)
		return 0, fmt.Errorf(variables.GrpcRecievError, err)
	}
	return grpcResponse.Value, nil
}
