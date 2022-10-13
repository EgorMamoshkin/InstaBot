package auth

type AuthServer interface {
	StartLS() error
}
type UserAccess struct {
	UserID int    `json:"user_id"`
	Token  string `json:"access_token"`
}
