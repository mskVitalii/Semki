package dto

type LoginRequest struct {
	Email        string `form:"email" json:"email" binding:"required" example:"msk.vitaly@gmail.com"`
	Password     string `form:"password" json:"password" binding:"required" example:"defaultPassword"`
	Organization string `form:"organization" json:"organization" binding:"required" example:"Staffbase"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest also delete refresh_token from redis
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UnauthorizedResponse struct {
	Message string `json:"message"`
}

type SuccessLogoutResponse struct {
	Message string `json:"message"`
}
