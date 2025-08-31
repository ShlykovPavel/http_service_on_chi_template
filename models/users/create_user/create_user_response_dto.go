package create_user

import (
	resp "github.com/ShlykovPavel/http_service_on_chi_template/internal/lib/api/response"
)

// CreateUserResponse Структура ответа на запрос
type CreateUserResponse struct {
	resp.Response
	UserID int64 `json:"id"`
}
