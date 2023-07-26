package controllers

import (
	"database/sql"
	_ "embed"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/m1/go-generate-password/generator"
	"log"
	"net/http"
	"sokoboxes-duo-api-orders/models"
	ordersRepository "sokoboxes-duo-api-orders/repository/orders"

	paypalRepository "sokoboxes-duo-api-orders/repository/paypal"
	sendgridRepository "sokoboxes-duo-api-orders/repository/sendgrid"
	"sokoboxes-duo-api-orders/utils"
	"strings"
)

type OrdersController struct {
	InProduction   bool
	PaypalClientId string
	PaypalSecret   string
	SendgridApiKey string
}

func (c *OrdersController) makePaypalRepository() paypalRepository.PaypalRestRepository {
	return paypalRepository.PaypalRestRepository{
		InProduction: c.InProduction,
		ClientId:     c.PaypalClientId,
		Secret:       c.PaypalSecret,
	}
}

func (c *OrdersController) makeSendgridRepository() sendgridRepository.SendgridApiRepository {
	return sendgridRepository.SendgridApiRepository{
		InProduction: c.InProduction,
		ApiKey:       c.SendgridApiKey,
	}
}

func (c *OrdersController) CreateOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		product := chi.URLParam(r, "product")
		if product != "commercial-levels" {
			w.WriteHeader(http.StatusBadRequest)
			utils.WriteJSON(w, map[string]string{"error": "Invalid product"})
			return
		}

		// TODO: Refactor later to retrieve from repository for better scalability and maintenance.
		priceCurrencyCode := "USD"
		priceValue := "12.00"

		paypalRepository := c.makePaypalRepository()

		accessToken, err := paypalRepository.GetAccessToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.WriteJSON(w, map[string]string{"error": err.Error()})
			return
		}

		order, err := paypalRepository.CreateOrder(product, priceCurrencyCode, priceValue, accessToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.WriteJSON(w, map[string]string{"error": err.Error()})
			return
		}

		createOrderResponse := models.CreateOrderResponse{
			IdOrder: order.Id,
		}
		utils.WriteJSON(w, createOrderResponse)
	}
}

func (c *OrdersController) generatePassword(length uint) (string, error) {
	config := generator.Config{
		Length:                     length,
		IncludeSymbols:             false,
		IncludeNumbers:             true,
		IncludeLowercaseLetters:    true,
		IncludeUppercaseLetters:    false,
		ExcludeSimilarCharacters:   true,
		ExcludeAmbiguousCharacters: true,
	}
	g, err := generator.New(&config)

	if err != nil {
		return "", err
	}

	pwd, err := g.Generate()
	if err != nil {
		return "", err
	}
	return *pwd, nil
}

func (c *OrdersController) CaptureOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idOrder := chi.URLParam(r, "idOrder")

		paypalRepository := c.makePaypalRepository()
		sendgridRepository := c.makeSendgridRepository()
		ordersRepository := ordersRepository.OrdersPgsqlRepository{}

		accessToken, err := paypalRepository.GetAccessToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.WriteJSON(w, map[string]string{"error": err.Error()})
			return
		}

		err, errorResponse, order := paypalRepository.CaptureOrder(idOrder, accessToken)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			utils.WriteJSON(w, map[string]string{"error": err.Error()})
			return
		}

		if errorResponse != nil {
			utils.WriteJSON(w, *errorResponse)
			return
		}

		code, err := c.generatePassword(30)
		if err != nil {
			log.Println(err.Error())
			uuidWithHyphen := uuid.New()
			code = strings.ToLower(order.Id + strings.Replace(uuidWithHyphen.String(), "-", "", -1))
			code = code[0:30]
		}

		log.Println("Generated Code:", code)

		err = ordersRepository.InsertOrders(order.Id, order.Status, order.Payer.EmailAddress, code, db)
		if err != nil {
			log.Println(err.Error())
			log.Println("Was not possible to insert the order Id:", order.Id, "Status:", order.Status, "E-mail:", order.Payer.EmailAddress)
		}

		fromName := "Sokoboxes Duo"
		fromEmail := "duo@sokoboxes.com"
		subject := "Your Sokoboxes Duo order"
		toName := ""
		toEmail := order.Payer.EmailAddress
		plainTextContent := `Sokoboxes Duo

Thank you for your order.
This is your code to unlock the 50 additional levels:

` + code + `

The code was automatically applied.

If you don't see the levels please refresh or input it in the Choose level section.

Carlos Montiers Aguilera

https://sokoboxes.com/duo
`

		htmlContent := `
<div style="color:#555555;margin:0;padding:0;height:auto;">
   <table cellpadding="0" cellspacing="0" border="0" width="100%" style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
      <tbody>
         <tr>
            <td align="left" valign="bottom" style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
               <img src="https://sokoboxes.com/duo/assets/img-sokoboxes-duo-logo.png" alt="Sokoboxes Duo"/>
            </td>
         </tr>
      </tbody>
   </table>
   <table style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;margin:5px 0;width:100%">
      <tbody>
         <tr>
            <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
               <h1 style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-family:Arial,sans-serif;font-size:14px;font-size:18px;font-weight:bold;margin:20px 0px 5px 0px;color:#555555">
                  Thank you for your order.
               </h1>
            </td>
         </tr>
         <tr>
            <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
               <table style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
                  <tbody>
                     <tr>
                        <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
                           <span style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;display:block;color:#555555">
                              <p style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;color:#555555">
                                 This is your code to unlock the 50 additional levels:
                              </p>
                           </span>
                           <table style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;padding:10px;text-align:center;margin:10px 0;margin:0!important;width:100%;width:100%!important;background:#f5f9fc;border:1px solid #b25d1b">
                              <tbody>
                                 <tr>
                                    <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
                                       <span style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;color:#555555">
                                          <h2 style="vertical-align:top;letter-spacing:normal;border-collapse:collapse;border-spacing:0;font-family:Arial,sans-serif;font-size:14px;font-size:16px;line-height:130%;line-height:18px;margin:10px 0 5px;margin:0;font-weight:bold;font-weight:normal;padding:15px 10px;text-align:center;color:#555555">
                                              ` + code + `
                                          </h2>
                                       </span>
                                    </td>
                                 </tr>
                              </tbody>
                           </table>
                        </td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
         <tr>
            <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
               <span style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;display:block;color:#555555">
                  <p style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;color:#555555">
                  The code was automatically applied.
                  </p>
               </span>
            </td>
         </tr>
         <tr>
            <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
               <span style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;display:block;color:#555555">
                  <p style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;padding:0;color:#555555">
                  If you don't see the levels please refresh or input it in the Choose level section.
                  </p>
               </span>
            </td>
         </tr>
      </tbody>
   </table>
   <table style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;width:100%;margin-top:20px;border-top:2px solid #b25d1b">
      <tbody>
         <tr>
            <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
               <table style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
                  <tbody>
                     <tr>
                        <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
                           <p style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;color:#555555;margin:5px 0;padding:0">
                           Carlos Montiers Aguilera
                           </p>
                        </td>
                     </tr>
                     <tr>
                        <td style="font-family:Arial,sans-serif;font-size:14px;border-spacing:0;border-collapse:collapse;line-height:130%;letter-spacing:normal;vertical-align:top">
                           <p style="vertical-align:top;letter-spacing:normal;line-height:130%;border-collapse:collapse;border-spacing:0;font-size:14px;font-family:Arial,sans-serif;font-weight:normal;color:#555555;margin:5px 0;padding:0">
                           <a href="https://sokoboxes.com/duo">https://sokoboxes.com/duo</a>
                           </p>
                        </td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
      </tbody>
   </table>
</div>
`
		err = sendgridRepository.SendEmail(
			fromName, fromEmail, subject, toName, toEmail, plainTextContent, htmlContent)
		if err != nil {
			log.Print(err.Error())
		}

		captureOrderResponse := models.CaptureOrderResponse{
			Code: code,
		}
		utils.WriteJSON(w, captureOrderResponse)
	}
}
