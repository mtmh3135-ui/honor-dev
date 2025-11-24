package models

type Activity struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Action      string `json:"action"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}
