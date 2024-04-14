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

func (r *BannerRepository) GetBanners(featureID int64, tagIDs []int64, limit, offset int64) ([]models.Banner, error) {
	query := `
		SELECT b.id, b.is_active, b.created_at, v.data, array_agg(bt.tag_id)
		FROM banners b
		INNER JOIN versions v ON b.id = v.banner_id
		INNER JOIN banner_tag bt ON b.id = bt.banner_id
		WHERE b.feature_id = $1 AND bt.tag_id = ANY($2)
		GROUP BY b.id, v.data
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(query, featureID, pq.Array(tagIDs), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []models.Banner
	for rows.Next() {
		var banner models.Banner
		err := rows.Scan(&banner.BannerID, &banner.IsActive, &banner.CreatedAt, &banner.Content, pq.Array(&banner.TagIDs))
		if err != nil {
			return nil, err
		}
		banners = append(banners, banner)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return banners, nil
}

func (r *BannerRepository) UserBanner(tagID int64, featureID int64, useLastRevision bool) (*models.Banner, error) {
	query := `
        SELECT b.id, b.is_active, b.created_at, v.data, json_build_object('tag_ids', json_agg(bt.tag_id)) as tag_ids
        FROM banners b
        INNER JOIN versions v ON b.id = v.banner_id
        INNER JOIN banner_tag bt ON b.id = bt.banner_id
        WHERE b.feature_id = $1 AND bt.tag_id = $2
    `
	// If useLastRevision is true, add additional condition to select the latest revision
	if useLastRevision {
		query += ` AND v.updated_at = (SELECT MAX(v2.updated_at) FROM versions v2 WHERE v2.banner_id = b.id) `
	}
	query += `
        GROUP BY b.id, b.is_active, b.created_at, v.data, v.updated_at
        ORDER BY v.updated_at DESC
        LIMIT 1
    `

	row := r.db.QueryRow(query, featureID, tagID)

	var banner models.Banner
	err := row.Scan(&banner.BannerID, &banner.IsActive, &banner.CreatedAt, &banner.Content, pq.Array(&banner.TagIDs))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &banner, nil
}
