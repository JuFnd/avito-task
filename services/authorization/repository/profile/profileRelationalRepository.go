package profile

import (
	"avito-track/pkg/models"
	"avito-track/pkg/variables"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/stdlib"
)

type ProfileRelationalRepository struct {
	db *sql.DB
}

func GetProfileRepository(configDatabase *variables.RelationalDataBaseConfig, logger *slog.Logger) (*ProfileRelationalRepository, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		configDatabase.User, configDatabase.DbName, configDatabase.Password, configDatabase.Host, configDatabase.Port, configDatabase.Sslmode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error(variables.SqlOpenError, err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Error(variables.SqlPingError, err.Error())
		return nil, err
	}

	db.SetMaxOpenConns(configDatabase.MaxOpenConns)

	profileDb := ProfileRelationalRepository{
		db: db,
	}

	errs := make(chan error)
	go func() {
		errs <- profileDb.pingDb(configDatabase.Timer, logger)
	}()

	if err := <-errs; err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return &profileDb, nil
}

func (repository *ProfileRelationalRepository) pingDb(timer uint32, logger *slog.Logger) error {
	var err error
	var retries int

	for retries < variables.MaxRetries {
		err = repository.db.Ping()
		if err == nil {
			return nil
		}

		retries++
		logger.Error(variables.SqlPingError, err.Error())
		time.Sleep(time.Duration(timer) * time.Second)
	}

	logger.Error(variables.SqlMaxPingRetriesError, err)
	return fmt.Errorf(variables.SqlMaxPingRetriesError, err.Error())
}

func (repository *ProfileRelationalRepository) CreateUser(login string, password []byte) error {
	_, err := repository.db.Exec(
		`INSERT INTO password(value)
			   VALUES ($1)`, password)
	if err != nil {
		return fmt.Errorf(variables.SqlProfileCreateError, err)
	}

	_, errProfile := repository.db.Exec(
		`INSERT INTO profile(login, password_id)
			   VALUES ($1, (SELECT id FROM password WHERE value = $2 LIMIT 1))`,
		login, password)
	if errProfile != nil {
		return fmt.Errorf(variables.SqlProfileCreateError, err)
	}

	_, errRole := repository.db.Exec(`INSERT INTO profile_role(profile_id, role_id)
                                             VALUES ((SELECT id FROM profile WHERE login = $1), $2)`, login, variables.UserRoleId)
	if errRole != nil {
		return fmt.Errorf(variables.SqlProfileCreateError, err)
	}
	return nil
}

func (repository *ProfileRelationalRepository) FindUser(login string) (bool, error) {
	userItem := &models.UserItem{}

	err := repository.db.QueryRow(
		`SELECT login FROM profile
			   WHERE login = $1`, login).Scan(&userItem.Login)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf(variables.ProfileNotFoundError, err.Error())
	}
	return true, nil
}

func (repository *ProfileRelationalRepository) GetUser(login string, password []byte) (*models.UserItem, bool, error) {
	userItem := &models.UserItem{}

	err := repository.db.QueryRow(
		`SELECT login FROM profile
			JOIN password ON profile.password_id = password.id
			WHERE profile.login = $1 AND password.value = $2`, login, password).Scan(&userItem.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, fmt.Errorf(variables.InvalidLoginOrPasswordError, ": %w", err)
		}
		return nil, false, fmt.Errorf(variables.ProfileNotFoundError, ": %w", err)
	}

	return userItem, true, nil
}

func (repository *ProfileRelationalRepository) GetUserProfileId(login string) (int64, error) {
	var userId int64

	err := repository.db.QueryRow("SELECT id FROM profile WHERE login = $1", login).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf(variables.ProfileIdNotFoundByLoginError, " %s", login)
		}
		return 0, fmt.Errorf(variables.FindProfileIdByLoginError, " %w", err)
	}
	return userId, nil
}

func (repository *ProfileRelationalRepository) GetUserRole(id int64) (string, error) {
	var role string

	err := repository.db.QueryRow(`SELECT role.value FROM profile
		JOIN profile_role ON profile.id = profile_role.profile_id
		JOIN role ON profile_role.role_id = role.id
		WHERE profile.id = $1`, id).Scan(&role)
	if err != nil {
		return "", fmt.Errorf(variables.ProfileRoleNotFoundByLoginError, " %w", err)
	}

	return role, nil
}
