package models

import "time"

type (
	Session struct {
		Login     string
		SID       string
		ExpiresAt time.Time
	}

	UserItem struct {
		Login string `json:"login"`
	}

	Banner struct {
		BannerID  int64                  `json:"banner_id"`
		TagIDs    []int64                `json:"tag_ids"`
		FeatureID int64                  `json:"feature_id"`
		Content   map[string]interface{} `json:"content"`
		IsActive  bool                   `json:"is_active"`
		CreatedAt string                 `json:"created_at"`
		UpdatedAt string                 `json:"updated_at"`
	}
)
