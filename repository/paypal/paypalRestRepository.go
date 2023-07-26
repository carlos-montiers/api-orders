package paypalRepository

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sokoboxes-duo-api-orders/models"
	"strings"
)

type PaypalRestRepository struct {
	InProduction bool
	ClientId     string
	Secret       string
	Product      string
}

func (p *PaypalRestRepository) getUrlBase() string {
	base := "https://api-m"
	if p.InProduction == false {
		base += ".sandbox"
	}
	base += ".paypal.com"
	return base
}

func (p *PaypalRestRepository) GetAccessToken() (string, error) {

	url := p.getUrlBase() + "/v1/oauth2/token"

	log.Println("url:", url)

	data := "grant_type=client_credentials"
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(p.ClientId, p.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Println(string(body))

	var jsonResponse models.PaypalTokenResponse
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", err
	}

	return jsonResponse.AccessToken, nil
}

func (p *PaypalRestRepository) CreateOrder(product string, priceCurrencyCode string, priceValue string, accessToken string) (models.PaypalCreateOrderResponse, error) {

	url := p.getUrlBase() + "/v2/checkout/orders"

	var orderResponse models.PaypalCreateOrderResponse

	var purchaseUnits []models.PurchaseUnit
	var item models.PurchaseUnit

	item = models.PurchaseUnit{
		Amount: models.Amount{
			CurrencyCode: priceCurrencyCode,
			Value:        priceValue,
		},
	}

	purchaseUnits = append(purchaseUnits, item)

	data := models.PaypalOrderData{
		Intent:        "CAPTURE",
		PurchaseUnits: purchaseUnits,
		ApplicationContext: models.ApplicationContext{
			ShippingPreference: "NO_SHIPPING",
			UserAction:         "PAY_NOW",
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return orderResponse, err
	}

	log.Println(string(jsonData))

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return orderResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return orderResponse, err
	}
	log.Printf("CreateOrder StatusCode: %d", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return orderResponse, err
	}

	log.Println(string(body))

	err = json.Unmarshal(body, &orderResponse)
	if err != nil {
		return orderResponse, err
	}

	return orderResponse, nil
}

func (p *PaypalRestRepository) CaptureOrder(idOrder string, accessToken string) (error, *models.PaypalErrorResponse, *models.PaypalCaptureOrderResponse) {

	url := p.getUrlBase() + "/v2/checkout/orders/" + idOrder + "/capture"

	var order models.PaypalCaptureOrderResponse
	var errorResponse models.PaypalErrorResponse

	req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		return err, nil, nil
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// req.Header.Set("PayPal-Mock-Response", "{\"mock_application_codes\": \"INSTRUMENT_DECLINED\"}")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, nil, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil, nil
	}

	log.Println(string(body))

	log.Printf("CaptureOrder StatusCode: %d", resp.StatusCode)
	success := resp.StatusCode == 200 || resp.StatusCode == 201

	if success {
		err = json.Unmarshal(body, &order)
		return err, nil, &order
	} else {
		err = json.Unmarshal(body, &errorResponse)
		return err, &errorResponse, nil
	}
}
