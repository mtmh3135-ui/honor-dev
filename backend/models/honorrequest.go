package models

type HonorRequest struct {
	ID           int     `json:"id"`
	Description  string  `json:"description"`
	CountedMonth int     `json:"counted_month"`
	CountedYear  int     `json:"counted_year"`
	Status       string  `json:"status"`
	CreatedBy    int64   `json:"created_by"`
	Username     string  `json:"username"`
	ApprovedLvl1 *string `json:"approved_lvl1"` // datetime bisa null → pointer
	ApprovedLvl2 *string `json:"approved_lvl2"`
	CreatedAt    string  `json:"created_at"`
	CancelledAt  *string `json:"cancelled_at"`
}
