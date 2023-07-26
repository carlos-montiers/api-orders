package driver

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

func ConnectDB(dataSourceName string) *sql.DB {

	// Delays in the attempts to connect to a serverless database.
	// If the database is idle, after the first connection attempt,
	// it begins the bootstrapping process.
	var sleeping = []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 0}

	var err error
	var db *sql.DB

	// Create connection
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal("open connection failed:", err.Error())
	}
	// Ping connection
	ctx := context.Background()

	for _, second := range sleeping {
		err = db.PingContext(ctx)
		if err == nil {
			// no error
			break
		}
		log.Println(err.Error())
		log.Printf("waiting %d seconds ...\n", second)
		time.Sleep(time.Duration(second) * time.Second)
	}
	if err != nil {
		os.Exit(1)
	}
	log.Printf("connected DB")

	return db
}
