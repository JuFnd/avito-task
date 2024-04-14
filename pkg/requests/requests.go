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
)
