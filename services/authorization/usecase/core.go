package usecase

import (
	"avito-track/pkg/models"
	"avito-track/pkg/util"
	"avito-track/pkg/variables"
	"avito-track/services/authorization/repository/profile"
	"avito-track/services/authorization/repository/session"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"
)

type IProfileRelationalRepository interface {
	CreateUser(login string, password []byte) error
	FindUser(login string) (bool, error)
	GetUser(login string, password []byte) (*models.UserItem, bool, error)
	GetUserProfileId(login string) (int64, error)
	GetUserRole(id int64) (string, error)
}

type ISessionCacheRepository interface {
	SaveSessionCache(ctx context.Context, createdSessionObject models.Session, logger *slog.Logger) (bool, error)
	GetSessionCache(ctx context.Context, sid string, logger *slog.Logger) (bool, error)
	DeleteSessionCache(ctx context.Context, sid string, logger *slog.Logger) (bool, error)
	GetUserLogin(ctx context.Context, sid string, logger *slog.Logger) (string, error)
}

type Core struct {
	sessions ISessionCacheRepository
	logger   *slog.Logger
	mutex    sync.RWMutex
	profiles IProfileRelationalRepository
}

func GetCore(profileConfig *variables.RelationalDataBaseConfig, sessionConfig *variables.CacheDataBaseConfig, logger *slog.Logger) (*Core, error) {
	sessionRepository, err := session.GetSessionRepository(sessionConfig, logger)
	if err != nil {
		logger.Error(variables.SessionRepositoryNotActiveError)
		return nil, err
	}

	profileRepository, err := profile.GetProfileRepository(profileConfig, logger)
	if err != nil {
		logger.Error(variables.ProfileRepositoryNotActiveError)
		return nil, err
	}

	core := Core{
		sessions: sessionRepository,
		logger:   logger.With(variables.ModuleLogger, variables.CoreModuleLogger),
		profiles: profileRepository,
	}

	return &core, nil
}

func (core *Core) CreateSession(ctx context.Context, login string) (models.Session, error) {
	sid := util.RandStringRunes(32)

	newSession := models.Session{
		Login:     login,
		SID:       sid,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}
	core.mutex.Lock()
	sessionAdded, err := core.sessions.SaveSessionCache(ctx, newSession, core.logger)
	defer core.mutex.Unlock()

	if !sessionAdded && err != nil {
		return models.Session{}, err
	}

	if !sessionAdded {
		return models.Session{}, nil
	}

	return newSession, nil
}

func (core *Core) KillSession(ctx context.Context, sid string) error {
	core.mutex.Lock()
	_, err := core.sessions.DeleteSessionCache(ctx, sid, core.logger)
	defer core.mutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (core *Core) FindActiveSession(ctx context.Context, sid string) (bool, error) {
	core.mutex.RLock()
	found, err := core.sessions.GetSessionCache(ctx, sid, core.logger)
	defer core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, nil
}

func (core *Core) CreateUserAccount(login string, password string) error {
	matched, err := regexp.MatchString(variables.LoginRegexp, login)
	if err != nil {
		core.logger.Error(variables.StatusInternalServerError, err.Error())
		return fmt.Errorf(variables.StatusInternalServerError, " %w", err)
	}
	if !matched {
		core.logger.Error(variables.InvalidLoginOrPasswordError)
		return fmt.Errorf(variables.InvalidLoginOrPasswordError)
	}

	hashPassword := util.HashPassword(password)
	err = core.profiles.CreateUser(login, hashPassword)
	if err != nil {
		core.logger.Error(variables.CreateProfileError, err.Error())
		return err
	}

	return nil
}

func (core *Core) FindUserByLogin(login string) (bool, error) {
	found, err := core.profiles.FindUser(login)
	if err != nil {
		core.logger.Error(variables.ProfileNotFoundError, err.Error())
		return false, err
	}

	return found, nil
}

func (core *Core) FindUserAccount(login string, password string) (*models.UserItem, bool, error) {
	hashPassword := util.HashPassword(password)
	user, found, err := core.profiles.GetUser(login, hashPassword)
	if err != nil {
		core.logger.Error(variables.ProfileNotFoundError, err.Error())
		return nil, false, err
	}
	return user, found, nil
}

func (core *Core) GetUserId(ctx context.Context, sid string) (int64, error) {
	login, err := core.sessions.GetUserLogin(ctx, sid, core.logger)
	if err != nil {
		return 0, err
	}

	id, err := core.profiles.GetUserProfileId(login)
	if err != nil {
		core.logger.Error(variables.GetProfileError, " id: %v", err)
		return 0, err
	}
	return id, nil
}

func (core *Core) GetUserRole(ctx context.Context, id int64) (string, error) {
	role, err := core.profiles.GetUserRole(id)
	if err != nil {
		core.logger.Error(variables.GetProfileRoleError, err.Error())
		return "", fmt.Errorf(variables.GetProfileRoleError, " %w", err)
	}

	return role, nil
}
