package models

type Payer struct {
	EmailAddress string `json:"email_address"`
}

type PaypalCreateOrderResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type PaypalGetOrderResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	Payer  Payer  `json:"payer"`
}

type PaypalCaptureOrderResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	Payer  Payer  `json:"payer"`
}

type PaypalErrorDetail struct {
	Issue string `json:"issue"`
}

type PaypalErrorResponse struct {
	Name string `json:"name"`
	Details []PaypalErrorDetail `json:"details"`
}
