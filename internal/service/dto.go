package service

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenInfo struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	ExpiresIn        int64     `json:"expires_in"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type UpdateProfileRequest struct {
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Disabled bool   `json:"disabled"`
}

type AuthorizeRequest struct {
	ClientID     string `json:"client_id" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required,url"`
	ResponseType string `json:"response_type" binding:"required,oneof=code"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
}

type TokenExchangeRequest struct {
	GrantType    string `json:"grant_type" binding:"required,oneof=authorization_code refresh_token"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}
