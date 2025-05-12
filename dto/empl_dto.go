package dto

type EmplLoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type DeliveryTransferInput struct {
	Barcode  string `json:"barcode" validate:"required"`
	SetterKu string `json:"setter_ku" validate:"required"`
	SetterEn string `json:"setter_en" validate:"required"`
	SetterAr string `json:"setter_ar" validate:"required"`
}
