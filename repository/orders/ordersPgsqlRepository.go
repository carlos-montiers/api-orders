package ordersRepository

import (
	"database/sql"
)

type OrdersPgsqlRepository struct {
}

func (o *OrdersPgsqlRepository) InsertOrders(idOrder string, status string, payerEmailAddress string, code string, db *sql.DB) error {

	_, err := db.Exec("INSERT INTO orders (id, status, payer_email_address, code) VALUES ($1, $2, $3, $4)",
		idOrder, status, payerEmailAddress, code,
	)
	return err
}
