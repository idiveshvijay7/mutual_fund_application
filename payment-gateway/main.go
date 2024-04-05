package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var errorRate = 0.1

func init() {
	rt := os.Getenv("ERROR_RATE")
	if rt != "" {
		v, err := strconv.ParseFloat(rt, 64)
		if err == nil {
			errorRate = v
		}
	}
}

func main() {
	godotenv.Load()
	// Initialize the database
	db, err := initializeDatabase()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	// Create a new instance of the App
	app := NewApp(db)

	// Run the application
	app.Run()
}

func initializeDatabase() (*sql.DB, error) {
	// Open the SQLite database file
	log.Default().Println("Initialising database...")
	db, err := sql.Open("sqlite3", "payments.db")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Create the payment table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT,
		account_number TEXT,
		ifsc_code TEXT,
		amount INTEGER,
		redirect_url TEXT,
		status TEXT DEFAULT 'Created',
		UTR TEXT,
		date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

type App struct {
	db *sql.DB
	// base url for the payment gateway
	baseURL string
}

func NewApp(db *sql.DB) *App {
	return &App{
		db: db,
		// baseurl from the environment variable
		baseURL: os.Getenv("BASE_URL"),
	}
}

func (a *App) Run() {
	// Start the web server
	mux := http.NewServeMux()
	mux.HandleFunc("POST /payment", randomFailureMiddleware(a.generatePaymentLinkHandler))
	mux.HandleFunc("GET /payment/{id}", randomFailureMiddleware(a.getPayment))
	mux.HandleFunc("GET /payment/pg/{id}", randomFailureMiddleware(a.paymentExecuteHandler))
	mux.HandleFunc("GET /payment/callback/{id}", a.paymentCallbackHandler)
	log.Default().Println("Server started at :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

type PaymentRequest struct {
	ID            string  `json:"id"`
	AccountNumber string  `json:"accountNumber"`
	IfscCode      string  `json:"ifscCode"`
	Amount        int64   `json:"amount"`
	RedirectUrl   string  `json:"redirectUrl"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"createdAt"`
	Utr           *string `json:"utr"`
}

func (a *App) generatePaymentLinkHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a PaymentRequest struct
	var paymentReq PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&paymentReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate the payment link using the bank account number and IFSC code
	paymentLink, err := a.generatePaymentLink(paymentReq)
	if err != nil {
		http.Error(w, "Error generating payment link", http.StatusInternalServerError)
		return
	}

	// Return the payment link in json response
	resp := map[string]interface{}{
		"paymentLink": paymentLink,
		"success":     true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *App) generatePaymentLink(req PaymentRequest) (string, error) {

	// Generate a UUID for the transaction
	transactionID := uuid.New().String()

	// store the details in a sql database
	_, err := a.db.Exec(`INSERT INTO payments (uuid, account_number, ifsc_code, amount, redirect_url, status) VALUES (?, ?, ?, ?, ?, ?)`, transactionID, req.AccountNumber, req.IfscCode, req.Amount, req.RedirectUrl, "Created")
	if err != nil {
		return "", fmt.Errorf("error inserting payment details into database: %w", err)
	}

	// Construct the payment URL
	paymentLink := a.baseURL + "/payment/pg/" + transactionID

	return paymentLink, nil
}

// handler function to get the payment details as json
func (a *App) getPayment(w http.ResponseWriter, r *http.Request) {
	// Get the transaction ID from the URL path
	transactionID := r.PathValue("id")

	// Query the database to get the payment details
	var payment PaymentRequest
	err := a.db.QueryRow("SELECT uuid, account_number, ifsc_code, amount, status, date, redirect_url, utr FROM payments WHERE uuid = ?", transactionID).Scan(&payment.ID, &payment.AccountNumber, &payment.IfscCode, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.RedirectUrl, &payment.Utr)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	// Return the payment details in json response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (a *App) paymentExecuteHandler(w http.ResponseWriter, r *http.Request) {
	// Get the transaction ID from the URL path
	transactionID := r.PathValue("id")

	// Query the database to get the payment status
	var status string
	err := a.db.QueryRow("SELECT status FROM payments WHERE uuid = ?", transactionID).Scan(&status)
	if err != nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	// Render the payment status page
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Payment Status</title>
		</head>
		<body>
			<h1>Payment Status</h1>
			<p>Transaction ID: %s</p>
			<p>Status: %s</p>
			<button onclick="window.location.href = '/payment/callback/%s?status=success';">Success</button>
			<button onclick="window.location.href = '/payment/callback/%s?status=failed';">Failed</button>
		</body>
		</html>
	`, transactionID, status, transactionID, transactionID)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (a *App) paymentCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the transaction ID from the URL path
	transactionID := r.PathValue("id")
	status := r.URL.Query().Get("status")

	if status != "success" && status != "failed" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	var dbStatus string
	a.db.QueryRow("SELECT status FROM payments WHERE uuid = ?", transactionID).Scan(&dbStatus)

	utr := ""
	if status == "success" {
		dbStatus = "Success"

		utr = fmt.Sprintf("ABCDBANK%s", transactionID)
	} else {
		dbStatus = "Failed"
	}

	// Update the payment status to "success" in the database
	_, err := a.db.Exec("UPDATE payments SET status = ?, utr = ? WHERE uuid = ?", dbStatus, utr, transactionID)
	if err != nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	var redirectUrl string
	a.db.QueryRow("SELECT redirect_url FROM payments WHERE uuid = ?", transactionID).Scan(&redirectUrl)

	random := rand.Float64()

	// If the random number is less than the failure rate, fail the request
	if random < errorRate {
		http.Error(w, "Request failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func randomFailureMiddleware(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	next := http.HandlerFunc(f)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a random number between 0 and 1
		random := rand.Float64()

		// If the random number is less than the failure rate, fail the request
		if random < errorRate {
			http.Error(w, "Request failed", http.StatusInternalServerError)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
