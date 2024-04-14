package delivery_grpc

import (
	"avito-track/configs"
	"avito-track/pkg/variables"
	pbAuth "avito-track/services/authorization/proto/authorization"
	"avito-track/services/authorization/repository/profile"
	"avito-track/services/authorization/repository/session"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type authorizationGrpc struct {
	grpcServer *grpc.Server
	logger     *slog.Logger
}

type authorizationGrpcServer struct {
	pbAuth.UnimplementedAuthorizationServer
	profileRepository *profile.ProfileRelationalRepository
	sessionRepository *session.SessionCacheRepository
	logger            *slog.Logger
}

func NewServer(configRelational *variables.RelationalDataBaseConfig, configSession *variables.CacheDataBaseConfig, logger *slog.Logger) (*authorizationGrpc, error) {
	session, err := session.GetSessionRepository(configSession, logger)

	if err != nil {
		logger.Error(variables.SessionRepositoryNotActiveError)
		return nil, fmt.Errorf(variables.GrpcListenAndServeError, ": %w", err)
	}

	users, err := profile.GetProfileRepository(configRelational, logger)
	if err != nil {
		logger.Error(variables.ProfileRepositoryNotActiveError)
		return nil, fmt.Errorf(variables.GrpcListenAndServeError, ": %w", err)
	}

	grpcServer := grpc.NewServer()
	pbAuth.RegisterAuthorizationServer(grpcServer, &authorizationGrpcServer{
		logger:            logger,
		sessionRepository: session,
		profileRepository: users,
	})

	return &authorizationGrpc{grpcServer: grpcServer, logger: logger}, nil
}

func (server *authorizationGrpc) ListenAndServeGrpc() error {
	grpcConfig, err := configs.ReadGrpcConfig()
	if err != nil {
		server.logger.Error(variables.ReadGrpcConfigError, ": %v", err)
		return fmt.Errorf(variables.GrpcListenAndServeError, ": %w", err)
	}

	lis, err := net.Listen(grpcConfig.ConnectionType, ":"+grpcConfig.Port)
	if err != nil {
		server.logger.Error(variables.GrpcListenAndServeError, ": %v", err)
		return fmt.Errorf(variables.GrpcListenAndServeError, ": %w", err)
	}

	if err := server.grpcServer.Serve(lis); err != nil {
		server.logger.Error(variables.GrpcListenAndServeError, ": %v", err)
		return fmt.Errorf(variables.GrpcListenAndServeError, ": %w", err)
	}

	return nil
}

func (server *authorizationGrpcServer) GetId(ctx context.Context, req *pbAuth.FindIdRequest) (*pbAuth.FindIdResponse, error) {
	login, err := server.sessionRepository.GetUserLogin(ctx, req.Sid, server.logger)
	if err != nil {
		return nil, err
	}

	id, err := server.profileRepository.GetUserProfileId(login)
	if err != nil {
		server.logger.Error(variables.ProfileNotFoundError, ": %v", err)
		return nil, err
	}
	return &pbAuth.FindIdResponse{
		Value: id,
	}, nil
}

func (server *authorizationGrpcServer) GetRole(ctx context.Context, req *pbAuth.RoleRequest) (*pbAuth.RoleResponse, error) {
	role, err := server.profileRepository.GetUserRole(req.Id)
	if err != nil {
		server.logger.Error(variables.GetProfileRoleError, ": %v", err)
		return nil, err
	}

	return &pbAuth.RoleResponse{
		Role: role,
	}, nil
}
