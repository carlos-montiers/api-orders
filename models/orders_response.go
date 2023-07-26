package models

type CreateOrderResponse struct {
	IdOrder string `json:"idOrder"`
}

type CaptureOrderResponse struct {
	Code   string `json:"code"`
}
