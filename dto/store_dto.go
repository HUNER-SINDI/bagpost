package dto

type StoreLoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type DeliveriesInsert struct {
	Barcode       string `json:"barcode" validate:"required"`
	CustomerPhone string `json:"customer_phone" validate:"required"`
	Note          string `json:"note"`

	ToCityKu string `json:"to_city_ku" validate:"required"`
	ToCityEn string `json:"to_city_en" validate:"required"`
	ToCityAr string `json:"to_city_ar" validate:"required"`

	ToSubCityKu string `json:"to_subcity_ku" validate:"required"`
	ToSubCityEn string `json:"to_subcity_en" validate:"required"`
	ToSubCityAr string `json:"to_subcity_ar" validate:"required"`

	ToSpecificLocation string `json:"to_specific_location"`
	Price              int32  `json:"price" validate:"gte=0"`
	Fee                int32  `json:"fdelivery_fee" validate:"required"`
	Total              int32  `json:"total_price" validate:"required"`
}
