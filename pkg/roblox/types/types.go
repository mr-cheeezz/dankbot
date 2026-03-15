package types

type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}
