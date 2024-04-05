package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var errorRate = 0.1

// 60 seconds
var navValueUpdateRate = 60

var processOrderRate = 5

func init() {
	rt := os.Getenv("ERROR_RATE")
	if rt != "" {
		v, err := strconv.ParseFloat(rt, 64)
		if err == nil {
			errorRate = v
		}
	}
	rt = os.Getenv("NAV_UPDATE_RATE")
	if rt != "" {
		v, err := strconv.Atoi(rt)
		if err == nil {
			navValueUpdateRate = v
		}
	}

	rt = os.Getenv("PROCESS_ORDER_RATE")
	if rt != "" {
		v, err := strconv.Atoi(rt)
		if err == nil {
			processOrderRate = v
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

	// run updateMarketValue every 1 minute
	go runNavUpdateCache()

	// Create a new instance of the App
	app := NewApp(db)

	// Run the application
	app.Run()
}

func runNavUpdateCache() {
	updateMarketValue()
	// run nav update every 1 minute
	ticker := time.NewTicker(time.Duration(navValueUpdateRate) * time.Second)
	go func() {
		for range ticker.C {
			updateMarketValue()
		}
	}()
}

func initializeDatabase() (*sql.DB, error) {
	// Open the SQLite database file
	log.Default().Println("Initialising database...")
	db, err := sql.Open("sqlite3", "orders.db")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Create the payment table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT,
		fund TEXT,
		amount FLOAT,
		units FLOAT,
		price_per_unit FLOAT,
		status TEXT DEFAULT 'Submitted',
		payment_id TEXT,
		submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		succeeded_at TIMESTAMP,
		failed_at TIMESTAMP
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
	baseURL           string
	paymentGatewayUrl string
}

func NewApp(db *sql.DB) *App {
	return &App{
		db: db,
		// baseurl from the environment variable
		baseURL:           os.Getenv("BASE_URL"),
		paymentGatewayUrl: os.Getenv("PAYMENT_GATEWAY_URL"),
	}
}

func (a *App) Run() {
	// Start the web server
	mux := http.NewServeMux()
	mux.HandleFunc("POST /order", randomFailureMiddleware(a.createOrderHandler))
	mux.HandleFunc("GET /order/{id}", randomFailureMiddleware(a.getOrder))
	mux.HandleFunc("GET /market-value/{fund}", randomFailureMiddleware(a.fundNav))
	log.Default().Println("Server started at :8081")
	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}

type OrderRequest struct {
	ID           string  `json:"id"`
	Fund         string  `json:"fund"`
	Amount       float64 `json:"amount"`
	Units        float64 `json:"units"`
	PricePerUnit float64 `json:"pricePerUnit"`
	Status       string  `json:"status"`
	PaymentID    string  `json:"paymentID"`
	SubmittedAt  *string `json:"submittedAt"`
	SucceededAt  *string `json:"succeededAt"`
	FailedAt     *string `json:"failedAt"`
}

func (a *App) createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a PaymentRequest struct
	var req OrderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate the payment link using the bank account number and IFSC code
	order, err := a.createOrder(req)
	if err != nil {
		http.Error(w, "Error gcreating order", http.StatusInternalServerError)
		return
	}

	// Return the payment link in json response
	resp := map[string]interface{}{
		"data":    order,
		"success": true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *App) createOrder(req OrderRequest) (*OrderRequest, error) {
	// Generate a UUID for the transaction
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	// Insert the payment details into the database
	_, err = a.db.Exec("INSERT INTO orders (uuid, fund, amount, units, price_per_unit, status, payment_id) VALUES (?, ?, ?, ?, ?, ?, ?)", uuid.String(), req.Fund, req.Amount, 0, 0, "Submitted", req.PaymentID)
	if err != nil {
		return nil, err
	}

	// Return the generated UUID
	req.ID = uuid.String()
	req.Status = "Submitted"
	go a.processOrder(uuid.String())
	return &req, nil
}

// handler function to get the order details as json
func (a *App) getOrder(w http.ResponseWriter, r *http.Request) {
	// Get the transaction ID from the URL path
	transactionID := r.PathValue("id")

	// Query the database to get the payment status
	var order OrderRequest
	err := a.db.QueryRow("SELECT uuid, fund, amount, units, price_per_unit, status, payment_id, submitted_at, succeeded_at, failed_at FROM orders WHERE uuid = ?", transactionID).Scan(&order.ID, &order.Fund, &order.Amount, &order.Units, &order.PricePerUnit, &order.Status, &order.PaymentID, &order.SubmittedAt, &order.SucceededAt, &order.FailedAt)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Return the payment details in json response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// handler function to get fund nav details
func (a *App) fundNav(w http.ResponseWriter, r *http.Request) {
	// Get the fund name from the URL path
	fundName := r.PathValue("fund")

	// Query the database to get the fund details
	fundC.Lock()
	defer fundC.Unlock()
	fund, ok := fundC.Funds[fundName]
	if !ok {
		http.Error(w, "Fund not found", http.StatusNotFound)
		return
	}

	// Return the fund details in json response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fund)
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

func (a *App) processOrder(orderID string) error {
	log.Default().Println("Processing order:", orderID)
	// Query the database to get the order details
	var order OrderRequest
	err := a.db.QueryRow("SELECT uuid, fund, amount, units, price_per_unit, status, payment_id, submitted_at, succeeded_at, failed_at FROM orders WHERE uuid = ?", orderID).Scan(&order.ID, &order.Fund, &order.Amount, &order.Units, &order.PricePerUnit, &order.Status, &order.PaymentID, &order.SubmittedAt, &order.SucceededAt, &order.FailedAt)
	if err != nil {
		log.Default().Println("Error getting order details:", err)
		return err
	}

	// checking if the payment is already processed for agiven payment id with successful orders in db
	var st string
	orderAlreadyProcessed := false
	err = a.db.QueryRow("SELECT status FROM orders WHERE payment_id = ? AND status = 'Succeeded'", order.PaymentID).Scan(&st)
	if err == nil {
		log.Default().Println("Payment already processed for order:", orderID)
		orderAlreadyProcessed = true
		order.Status = "Failed"
	}

	if !orderAlreadyProcessed {
		// Check with the payment gateway if the payment is successful
		log.Default().Println("Checking payment status for order:", orderID)
		paymentStatus, err := a.checkPaymentStatus(order.PaymentID)
		if err != nil {
			log.Default().Println("Error checking payment status:", err)
			paymentStatus = "Couldn't fetch payment status"
		}
		log.Default().Println("Payment status for order:", orderID, "is", paymentStatus)

		// Update the order status based on the payment status
		log.Default().Println("Updating order status for order:", orderID)
		switch paymentStatus {
		case "Success":
			order.Status = "Succeeded"
		case "Failed":
			order.Status = "Failed"
		default:
			order.Status = "Failed"
		}
	}

	// Simulate processing the order
	time.Sleep(time.Duration(processOrderRate) * time.Second)

	// update order status
	fundC.Lock()
	defer fundC.Unlock()
	fund, ok := fundC.Funds[order.Fund]
	units := order.Amount / fund.MarketValue
	if order.Status == "Succeeded" && ok && units >= 1 {
		ppu := fund.MarketValue
		a.db.Exec("UPDATE orders SET units = ?, price_per_unit = ?, succeeded_at = CURRENT_TIMESTAMP WHERE uuid = ?", units, ppu, orderID)
	} else {
		a.db.Exec("UPDATE orders SET failed_at = CURRENT_TIMESTAMP WHERE uuid = ?", orderID)
	}

	log.Default().Println("Order status updated for order:", orderID, "to", order.Status)

	return nil
}

func (a *App) checkPaymentStatus(paymentID string) (string, error) {
	// Call the payment gateway API to get the payment status
	url := a.paymentGatewayUrl + "/payment/" + paymentID
	log.Default().Println("Checking payment status at:", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading payment response: %w", err)
	}

	// Parse the response body into a PaymentRequest struct
	var paymentReq PaymentRequest
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&paymentReq)
	if err != nil {
		return "", fmt.Errorf("error decoding payment response: %w. response %+v", err, string(body))
	}

	return paymentReq.Status, nil
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

type Fund struct {
	Name        string  `json:"name"`
	NavMin      float64 `json:"-"`
	NavMax      float64 `json:"-"`
	MarketValue float64 `json:"marketValue"`
}

type fundCache struct {
	Funds map[string]Fund
	sync.Mutex
}

var fundC = fundCache{
	Funds: map[string]Fund{
		"Arbitrage Fund 1": {Name: "Arbitrage Fund 1", NavMin: 10.0, NavMax: 20.0},
		"Arbitrage Fund 2": {Name: "Arbitrage Fund 2", NavMin: 10.0, NavMax: 30.0},
		"Arbitrage Fund 3": {Name: "Arbitrage Fund 3", NavMin: 100.0, NavMax: 200.0},
		"Arbitrage Fund 4": {Name: "Arbitrage Fund 4", NavMin: 50.0, NavMax: 60.0},
		"Balanced Fund 1":  {Name: "Balanced Fund 1", NavMin: 20.0, NavMax: 100.0},
		"Balanced Fund 2":  {Name: "Balanced Fund 2", NavMin: 10.0, NavMax: 200.0},
		"Balanced Fund 3":  {Name: "Balanced Fund 3", NavMin: 60.0, NavMax: 300.0},
		"Balanced Fund 4":  {Name: "Balanced Fund 4", NavMin: 100.0, NavMax: 400.0},
		"Balanced Fund 5":  {Name: "Balanced Fund 5", NavMin: 300.0, NavMax: 200.0},
		"Growth Fund 1":    {Name: "Growth Fund 1", NavMin: 50.0, NavMax: 100.0},
		"Growth Fund 2":    {Name: "Growth Fund 2", NavMin: 60.0, NavMax: 150.0},
		"Growth Fund 3":    {Name: "Growth Fund 3", NavMin: 70.0, NavMax: 200.0},
		"Growth Fund 4":    {Name: "Growth Fund 4", NavMin: 80.0, NavMax: 250.0},
		"Growth Fund 5":    {Name: "Growth Fund 5", NavMin: 90.0, NavMax: 300.0},
	},
}

// function to update the market value with random value but within NavMin and NavMax range
func updateMarketValue() {
	fundC.Lock()

	for k, fund := range fundC.Funds {
		fund.MarketValue = fund.NavMin + rand.Float64()*(fund.NavMax-fund.NavMin)
		fundC.Funds[k] = fund
	}

	fundC.Unlock()
}
