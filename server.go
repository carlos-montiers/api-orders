package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sokoboxes-duo-api-orders/controllers"
	"sokoboxes-duo-api-orders/driver"
	"sokoboxes-duo-api-orders/utils"
	"syscall"
	"time"
)

func main() {
	port := utils.GetEnv("PORT", false)
	if port == "" {
		port = "8080"
	}
	PRODUCTION := utils.GetBoolEnv("PRODUCTION")
	PAYPAL_CLIENT_ID := utils.GetEnv("PAYPAL_CLIENT_ID", true)
	PAYPAL_SECRET := utils.GetEnv("PAYPAL_SECRET", true)
	DATA_SOURCE_NAME := utils.GetEnv("DATA_SOURCE_NAME", true)
	SENDGRID_API_KEY := utils.GetEnv("SENDGRID_API_KEY", true)

	db := driver.ConnectDB(DATA_SOURCE_NAME)

	ordersController := controllers.OrdersController{
		InProduction:   PRODUCTION,
		PaypalClientId: PAYPAL_CLIENT_ID,
		PaypalSecret:   PAYPAL_SECRET,
		SendgridApiKey: SENDGRID_API_KEY,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	var allowedOrigins []string
	if PRODUCTION {
		allowedOrigins = []string{"https://sokoboxes.com"}
	} else {
		allowedOrigins = []string{"*"}
	}
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: false,
		Debug:            !PRODUCTION,
	}).Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, map[string]string{"svc": "orders"})
	})

	r.Post("/checkout/api/paypal/order/create/{product}", ordersController.CreateOrder(db))
	r.Post("/checkout/api/paypal/order/{idOrder}/capture", ordersController.CaptureOrder(db))

	r.Post("/key/{key}", func(w http.ResponseWriter, r *http.Request) {
		// key := chi.URLParam(r, "key")
		utils.WriteJSON(w, map[string]string{"status": "success"})
	})

	log.Printf("starting up on port %s", port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(r.ServeHTTP),
	}

	// Create channel to listen for signals.
	signalChan := make(chan os.Signal, 1)
	// SIGINT handles Ctrl+C locally.
	// SIGTERM handles Cloud Run termination signal.
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server.
	go func() {
		log.Printf("listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Receive output from signalChan.
	sig := <-signalChan
	log.Printf("%s signal caught", sig)

	// Timeout if waiting for connections to return idle.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add extra handling here to clean up resources, such as flushing logs and
	// closing any database or Redis connections.
	log.Printf("disconnecting DB")
	db.Close()

	// Gracefully shutdown the server by waiting on existing requests (except websockets).
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %+v", err)
	}
	log.Print("server exited")
}
