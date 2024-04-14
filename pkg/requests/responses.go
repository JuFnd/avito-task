package communication

type (
	SignupResponse struct {
		Login string `json:"login"`
	}

	AdvertItemResponse struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Date        string `json:"date"`
		Price       int64  `json:"price"`
		ImagePath   string `json:"image_path"`
		ProfileId   int64  `json:"profile_id"`
		IsAuthor    bool   `json:"is_author"`
	}
)
