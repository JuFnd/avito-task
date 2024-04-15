package repository

import (
	"avito-track/pkg/models"
	"avito-track/pkg/variables"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/lib/pq"
	"log/slog"
	"time"
)

type BannerRepository struct {
	db *sql.DB
}

func GetBannerRepository(configDatabase variables.RelationalDataBaseConfig, logger *slog.Logger) (*BannerRepository, error) {
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

	bannerRepository := &BannerRepository{db: db}

	errs := make(chan error)
	go func() {
		errs <- bannerRepository.pingDb(configDatabase.Timer, logger)
	}()

	if err := <-errs; err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return bannerRepository, nil
}

func (repository *BannerRepository) pingDb(timer uint32, logger *slog.Logger) error {
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

func (repository *BannerRepository) GetBanners(userRole string, featureID int64, tagIDs []int64, limit, offset int64) ([]models.Banner, error) {
	query := `
		SELECT b.id, b.feature_id, v.is_active, b.created_at, v.updated_at, v.data, array_agg(bt.tag_id)
		FROM banners b
		INNER JOIN versions v ON b.id = v.banner_id
		INNER JOIN banner_tag bt ON b.id = bt.banner_id
		WHERE b.feature_id = $1 AND bt.tag_id = ANY($2)
		GROUP BY b.id, v.data, v.is_active, v.updated_at
		LIMIT $3 OFFSET $4
	`
	rows, err := repository.db.Query(query, featureID, pq.Array(tagIDs), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []models.Banner
	for rows.Next() {
		var banner models.Banner
		err := rows.Scan(&banner.BannerID, &banner.FeatureID, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &banner.Content, pq.Array(&banner.TagIDs))
		if err != nil {
			return nil, err
		}
		if userRole == variables.UserRole[0] && banner.IsActive {
			banners = append(banners, banner)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return banners, nil
}

func (repository *BannerRepository) AddBanner(tagIds []int64, featureID int64, content string) error {
	tx, err := repository.db.Begin()
	if err != nil {
		return err
	}

	var bannerID int64
	err = tx.QueryRow("INSERT INTO banners (feature_id) VALUES ($1) RETURNING id", featureID).Scan(&bannerID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, tagID := range tagIds {
		_, err := tx.Exec("INSERT INTO banner_tag (banner_id, tag_id) VALUES ($1, $2)", bannerID, tagID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_, err = tx.Exec("INSERT INTO versions (banner_id, data) VALUES ($1, $2)", bannerID, content)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (repository *BannerRepository) UserBanner(tagID int64, featureID int64, useLastRevision bool) (*models.Banner, error) {
	var query string
	if useLastRevision {
		query = `
            SELECT b.id, b.feature_id, v.is_active, MAX(b.created_at), v.updated_at, v.data, array_agg(bt.tag_id)
            FROM banners b
            INNER JOIN versions v ON b.id = v.banner_id
            LEFT JOIN banner_tag bt ON b.id = bt.banner_id
            WHERE b.feature_id = $1 AND bt.tag_id = $2 AND v.is_active = TRUE
            GROUP BY b.id, b.feature_id, v.is_active, b.created_at, v.data, v.updated_at
            ORDER BY v.updated_at DESC
            LIMIT 1
        `
	} else {
		query = `
            SELECT b.id, b.feature_id, v.is_active, b.created_at, v.updated_at, v.data, array_agg(bt.tag_id)
            FROM banners b
            INNER JOIN versions v ON b.id = v.banner_id
            LEFT JOIN banner_tag bt ON b.id = bt.banner_id
            WHERE b.feature_id = $1 AND bt.tag_id = $2 AND v.is_active = TRUE AND v.updated_at <= NOW() - INTERVAL '5 MINUTES'
            GROUP BY b.id, v.is_active, b.created_at, v.data, v.updated_at
        `
	}

	row := repository.db.QueryRow(query, featureID, tagID)

	var banner models.Banner
	err := row.Scan(&banner.BannerID, &banner.FeatureID, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &banner.Content, pq.Array(&banner.TagIDs))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (repository *BannerRepository) UpdateBanner(id int64, tagIds []int64, featureID int64, content string) error {
	tx, err := repository.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE versions SET is_active = FALSE WHERE banner_id = $1 AND is_active = TRUE", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("INSERT INTO versions (banner_id, is_active, data) VALUES ($1, TRUE, $2)", id, content)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE banners SET feature_id = $1 WHERE id = $2", featureID, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM banner_tag WHERE banner_id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, tagID := range tagIds {
		_, err = tx.Exec("INSERT INTO banner_tag (banner_id, tag_id) VALUES ($1, $2)", id, tagID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (repository *BannerRepository) DeleteBanner(id int64) error {
	tx, err := repository.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM banners WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM versions WHERE banner_id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM banner_tag WHERE banner_id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
