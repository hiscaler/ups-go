package entity

// Token OAuth 访问令牌
type Token struct {
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	ClientID    string `json:"client_id"`
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"` // 秒数字符串
	Status      string `json:"status"`
}
