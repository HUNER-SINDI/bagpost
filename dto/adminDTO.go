package dto

type CreateWarehouseInput struct {
	Fullname string `json:"fullname" validate:"required"`
	Locname  string `json:"locname" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	IsActive bool   `json:"is_active"`
	Krd      string `json:"krd" validate:"required"`
	Ar       string `json:"ar" validate:"required"`
}

type UpdateWarehouseInput struct {
	Fullname string `json:"fullname" validate:"required"`
	Locname  string `json:"locname" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
	Krd      string `json:"krd" validate:"required"`
	Ar       string `json:"ar" validate:"required"`
}
