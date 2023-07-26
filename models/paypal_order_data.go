package models

type Amount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type PurchaseUnit struct {
	Amount Amount `json:"amount"`
}

type ApplicationContext struct {
	ShippingPreference string `json:"shipping_preference"`
	UserAction         string `json:"user_actionenum"`
}

type PaypalOrderData struct {
	Intent             string             `json:"intent"`
	PurchaseUnits      []PurchaseUnit     `json:"purchase_units"`
	ApplicationContext ApplicationContext `json:"application_context"`
}
