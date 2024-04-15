package communication

type (
	SigninRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	SignupRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	BannerRequest struct {
		TagIds    []int64 `json:"tag_ids"`
		FeatureId int64   `json:"feature_id"`
		Content   string  `json:"content"`
	}
)
