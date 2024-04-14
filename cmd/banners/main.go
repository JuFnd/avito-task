package main

import (
	"avito-track/configs"
	"avito-track/pkg/variables"
	"avito-track/services/banners/delivery"
	"avito-track/services/banners/repository"
	"avito-track/services/banners/usecase"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logFile, err := os.Create("banners.log")
	if err != nil {
		fmt.Println("Error creating log file")
		return
	}

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	bannersAppConfig, err := configs.ReadBannersAppConfig()
	if err != nil {
		logger.Error(variables.ReadAuthConfigError, err.Error())
		return
	}

	relationalDataBaseConfig, err := configs.ReadRelationalBannersDataBaseConfig()
	if err != nil {
		logger.Error(variables.ReadAuthSqlConfigError, err.Error())
		return
	}

	bannersRepository, err := repository.GetBannerRepository(*relationalDataBaseConfig, logger)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	grpcConfig, err := configs.ReadGrpcConfig()
	if err != nil {
		logger.Error(variables.ReadGrpcConfigError, err.Error())
		return
	}
	core := usecase.GetCore(*grpcConfig, bannersRepository, logger)

	api := delivery.GetApi(core, logger)

	err = api.ListenAndServe(bannersAppConfig)
	if err != nil {
		logger.Error(err.Error())
	}
}
