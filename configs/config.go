package configs

import (
	"avito-track/pkg/variables"
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"syscall"
)

func readYAMLFile[T any](filePath string) (*T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %w", err)
	}

	var config T
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML data: %w", err)
	}

	return &config, nil
}

func ParseFlagsAndReadYAMLFile[T any](fileName string, defaultFilePath string, flags *flag.FlagSet) (*T, error) {
	flag.Parse()
	var path string
	flag.StringVar(&path, fileName, defaultFilePath, "Путь к конфигу"+fileName)

	config, err := readYAMLFile[T](path)
	if ok := errors.As(err, &err); ok && err == syscall.ENOENT {
		return nil, fmt.Errorf("Failed to parse '%s' from provided path: %w", fileName, err)
	}

	return config, nil
}

func ReadAuthAppConfig() (*variables.AppConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.AppConfig]("auth_config_path", "../../configs/AuthorizationAppConfig.yml", flag.CommandLine)
}

func ReadGrpcConfig() (*variables.GrpcConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.GrpcConfig]("grpc_config_path", "../../configs/GrpcConfig.yml", flag.CommandLine)
}

func ReadBannersAppConfig() (*variables.AppConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.AppConfig]("market_config_path", "../../configs/BannersAppConfig.yml", flag.CommandLine)
}

func ReadRelationalAuthDataBaseConfig() (*variables.RelationalDataBaseConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.RelationalDataBaseConfig]("sql_config_auth_path", "../../configs/AuthorizationSqlDataBaseConfig.yml", flag.CommandLine)
}

func ReadRelationalBannersDataBaseConfig() (*variables.RelationalDataBaseConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.RelationalDataBaseConfig]("sql_config_films_path", "../../configs/BannersSqlDataBaseConfig.yml", flag.CommandLine)
}

func ReadCacheDatabaseConfig() (*variables.CacheDataBaseConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.CacheDataBaseConfig]("cache_config_path", "../../configs/AuthorizationCacheDataBaseConfig.yml", flag.CommandLine)
}
