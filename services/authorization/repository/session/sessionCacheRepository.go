package session

import (
	"avito-track/pkg/models"
	"avito-track/pkg/variables"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionCacheRepository struct {
	sessionRedisClient *redis.Client
}

func (sessionCacheRepository *SessionCacheRepository) reconnectRedis() error {
	err := sessionCacheRepository.sessionRedisClient.Close()
	if err != nil {
		return err
	}

	newClient := redis.NewClient(&redis.Options{
		Addr:     sessionCacheRepository.sessionRedisClient.Options().Addr,
		Password: sessionCacheRepository.sessionRedisClient.Options().Password,
		DB:       sessionCacheRepository.sessionRedisClient.Options().DB,
	})

	sessionCacheRepository.sessionRedisClient = newClient

	return nil
}

func (sessionCacheRepository *SessionCacheRepository) pingRedis(timer int, logger *slog.Logger) error {
	var pingErrString string
	var reconnectErrString string
	var retries int

	for retries < variables.MaxRetries {
		_, pingErr := sessionCacheRepository.sessionRedisClient.Ping(sessionCacheRepository.sessionRedisClient.Context()).Result()
		if pingErr == nil {
			return nil
		}
		pingErrString = pingErr.Error()

		reconnectErr := sessionCacheRepository.reconnectRedis()
		if reconnectErr == nil {
			return nil
		}
		reconnectErrString = reconnectErr.Error()

		retries++
		logger.Error(variables.AuthorizationCachePingRetryError, pingErr.Error(), reconnectErr.Error())
		time.Sleep(time.Duration(timer) * time.Second)
	}

	return fmt.Errorf(variables.AuthorizationCachePingMaxRetriesError, pingErrString, reconnectErrString)
}

func GetSessionRepository(sessionConfig *variables.CacheDataBaseConfig, logger *slog.Logger) (*SessionCacheRepository, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     sessionConfig.Host,
		Password: sessionConfig.Password,
		DB:       sessionConfig.DbNumber,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	sessionCacheRepository := &SessionCacheRepository{
		sessionRedisClient: redisClient,
	}

	errs := make(chan error)

	go func() {
		errs <- sessionCacheRepository.pingRedis(sessionConfig.Timer, logger)
	}()

	if err := <-errs; err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return sessionCacheRepository, nil
}

func (sessionCacheRepository *SessionCacheRepository) SaveSessionCache(ctx context.Context, createdSessionObject models.Session, logger *slog.Logger) (bool, error) {
	sessionCacheRepository.sessionRedisClient.Set(ctx, createdSessionObject.SID, createdSessionObject.Login, 24*time.Hour)

	sessionAdded, errCheck := sessionCacheRepository.GetSessionCache(ctx, createdSessionObject.SID, logger)

	if errCheck != nil {
		return false, errCheck
	}

	return sessionAdded, nil
}

func (sessionCacheRepository *SessionCacheRepository) GetSessionCache(ctx context.Context, sid string, logger *slog.Logger) (bool, error) {
	_, err := sessionCacheRepository.sessionRedisClient.Get(ctx, sid).Result()
	if err == redis.Nil {
		logger.Error(variables.SessionNotFoundError, sid)
		return false, nil
	}

	if err != nil {
		logger.Error(variables.StatusInternalServerError, err)
		return false, err
	}

	return true, nil
}

func (sessionCacheRepository *SessionCacheRepository) DeleteSessionCache(ctx context.Context, sid string, logger *slog.Logger) (bool, error) {
	_, err := sessionCacheRepository.sessionRedisClient.Del(ctx, sid).Result()
	if err != nil {
		logger.Error(variables.SessionRemoveError, err)
		return false, err
	}

	return true, nil
}

func (sessionCacheRepository *SessionCacheRepository) GetUserLogin(ctx context.Context, sid string, logger *slog.Logger) (string, error) {
	value, err := sessionCacheRepository.sessionRedisClient.Get(ctx, sid).Result()
	if err != nil {
		logger.Error(variables.SessionNotFoundError + sid)
		return "", err
	}

	return value, nil
}
