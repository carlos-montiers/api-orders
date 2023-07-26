package sendgridRepository

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendgridApiRepository struct {
	InProduction bool
	ApiKey       string
}

func (s *SendgridApiRepository) SendEmail(
	fromName string,
	fromEmail string,
	subject string,
	toName string,
	toEmail string,
	plainTextContent string,
	htmlContent string,
) error {

	from := mail.NewEmail(fromName, fromEmail)
	to := mail.NewEmail(toName, toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(s.ApiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}
	fmt.Println(response.StatusCode)
	fmt.Println(response.Body)
	fmt.Println(response.Headers)
	return nil
}
