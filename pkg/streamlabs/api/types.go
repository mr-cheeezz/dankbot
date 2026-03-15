package api

type SocketTokenResponse struct {
	SocketToken string `json:"socket_token,omitempty"`
	Token       string `json:"token,omitempty"`
}

type AlertRequest map[string]any

type UserResponse map[string]any

type PointsResponse map[string]any

type Donation struct {
	DonationID string  `json:"donation_id"`
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Message    string  `json:"message"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	CreatedAt  string  `json:"created_at"`
}

type DonationsResponse struct {
	Data         []Donation     `json:"data"`
	Pagination   map[string]any `json:"pagination"`
	Total        int            `json:"total"`
	Page         int            `json:"page"`
	PerPage      int            `json:"per_page"`
	CurrentPage  int            `json:"current_page"`
	LastPage     int            `json:"last_page"`
	NextPageURL  string         `json:"next_page_url"`
	PrevPageURL  string         `json:"prev_page_url"`
	FirstPageURL string         `json:"first_page_url"`
	LastPageURL  string         `json:"last_page_url"`
	From         int            `json:"from"`
	To           int            `json:"to"`
}
