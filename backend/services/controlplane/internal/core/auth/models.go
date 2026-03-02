package auth

// ExchangeRequest representa o payload de entrada para troca de token.
type ExchangeRequest struct {
	ExternalToken string `json:"external_token" binding:"required"`
}

// TokenResponse representa o payload de retorno contendo o nosso JWT interno.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}
