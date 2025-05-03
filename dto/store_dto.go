package dto

type StoreLoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type DeliveriesInsert struct {
	Barcode            string `json:"barcode" validate:"required"`
	CustomerPhone      string `json:"customer_phone" validate:"required"`
	Note               string `json:"note"`
	FromCity           string `json:"from_city" validate:"required"`
	ToCity             string `json:"to_city" validate:"required"`
	ToSubCity          string `json:"to_subcity" validate:"required"`
	ToSpecificLocation string `json:"to_specific_location" validate:"required"`
	Status             string `json:"status"`
	Price              int32  `json:"price" validate:"gte=0"`
	Fee                int32  `json:"fdelivery_fee" validate:"required"`
	Total              int32  `json:"total_price" validate:"required"`
}
