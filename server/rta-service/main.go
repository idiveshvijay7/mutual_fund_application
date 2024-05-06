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

	// Create the orders table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone_number TEXT,
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

	// Create the user table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone_number TEXT UNIQUE
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

	// Add this handler to your router or mux
	mux.HandleFunc("POST /execute-strategy-orders", randomFailureMiddleware(a.executeStrategyOrdersHandler))
	// Route for user login
	mux.HandleFunc("POST /login", randomFailureMiddleware(a.loginHandler))

	// Route for user signup
	mux.HandleFunc("POST /signup", randomFailureMiddleware(a.signupHandler))
	mux.HandleFunc("/aggregated-orders-by-phone", randomFailureMiddleware(a.getAggregatedOrdersByPhoneNumber))

	handler := allowCORS(mux)

	// Route for fetching user portfolio
	// mux.HandleFunc("/portfolio", randomFailureMiddleware(a.getPortfolioHandler))
	log.Default().Println("Server started at :8081")
	err := http.ListenAndServe(":8081", handler)
	if err != nil {
		log.Fatal(err)
	}
}

// CORS middleware // Add this handler to your router or mux
func allowCORS(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
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
	PhoneNumber  string  `json:"phoneNumber"`
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

func (a *App) executeStrategyOrdersHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the request body into a struct containing the necessary data
    var requestData struct {
        StrategyName string  `json:"strategyName"`
        Amount       float64 `json:"amount"`
        PaymentID    string  `json:"paymentID"`
        PhoneNumber  string  `json:"phoneNumber"`
    }

    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "Failed to decode request body", http.StatusBadRequest)
        return
    }

    // Create a channel to signal completion
    done := make(chan struct{})

    // Execute the strategy orders using the provided data in a goroutine
    go func() {
        if err := a.executeStrategyOrders(requestData.StrategyName, requestData.Amount, requestData.PaymentID, requestData.PhoneNumber); err != nil {
            // Handle any errors if needed
            fmt.Printf("Failed to execute strategy orders: %v\n", err)
        }
        // Signal completion
        done <- struct{}{}
    }()

    // Wait for the function to finish
    <-done

    // Respond with a success message
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Strategy orders executed successfully"))
}

func (a *App) executeStrategyOrders(strategyName string, amount float64, paymentID, phoneNumber string) error {
	// Retrieve strategy details based on the strategy name
	strategyFundsMap := convertStrategyJsonIntoMap()
	strategy, ok := strategyFundsMap[strategyName]
	if !ok {
		return fmt.Errorf("strategy '%s' not found", strategyName)
	}

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Iterate over each fund in the strategy and create an order for it
	for _, fund := range strategy {
		// Increment the wait group counter
		wg.Add(1)

		// Goroutine to create order for the current fund
		go func(fund Funds) {
			defer wg.Done()

			// Calculate the amount for the current fund based on its percentage in the strategy
			fundAmount := (amount * float64(fund.Percentage)) / 100

			// Create an order request for the current fund
			orderReq := OrderRequest{
				Fund:        fund.Name,
				Amount:      fundAmount,
				PaymentID:   paymentID,
				PhoneNumber: phoneNumber,
				// You may need to provide other required fields like PhoneNumber, etc.
			}

			// Create the order for the current fund
			_, err := a.createOrder(orderReq)
			if err != nil {
				// Handle error if order creation fails
				log.Printf("Error creating order for fund %s: %v\n", fund.Name, err)
				return
			}
		}(fund)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return nil
}

func (a *App) createOrder(req OrderRequest) (*OrderRequest, error) {
	// Generate a UUID for the transaction
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	// Insert the payment details into the database
	_, err = a.db.Exec("INSERT INTO orders (uuid, fund, amount, units, price_per_unit, status, payment_id, phone_number) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", uuid.String(), req.Fund, req.Amount, 0, 0, "Submitted", req.PaymentID, req.PhoneNumber)
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
	err := a.db.QueryRow("SELECT uuid, fund, amount, units, price_per_unit, status, payment_id, phone_number, submitted_at, succeeded_at, failed_at FROM orders WHERE uuid = ?", transactionID).Scan(&order.ID, &order.Fund, &order.Amount, &order.Units, &order.PricePerUnit, &order.Status, &order.PaymentID, &order.PhoneNumber, &order.SubmittedAt, &order.SucceededAt, &order.FailedAt)
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
		paymentStatus, err := a.retryCheckPaymentStatus(order.PaymentID, 2) // Retry 2 times
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
		a.db.Exec("UPDATE orders SET units = ?, price_per_unit = ?,  status = ?, succeeded_at = CURRENT_TIMESTAMP WHERE uuid = ?", units, ppu, "Succeeded", orderID)
	} else {
		a.db.Exec("UPDATE orders SET failed_at = CURRENT_TIMESTAMP, status = ? WHERE uuid = ?", "Failed", orderID)
	}

	log.Default().Println("Order status updated for order:", orderID, "to", order.Status)

	return nil
}

// RetryCheckPaymentStatus retries checking the payment status for the given payment ID for a specified number of times.
func (a *App) retryCheckPaymentStatus(paymentID string, retryCount int) (string, error) {
	for i := 0; i < retryCount; i++ {
		paymentStatus, err := a.checkPaymentStatus(paymentID)
		if err == nil {
			return paymentStatus, nil
		}
		log.Printf("Error checking payment status (attempt %d/%d): %v", i+1, retryCount, err)
		time.Sleep(1 * time.Second) // Add a delay between retries
	}
	return "", fmt.Errorf("unable to check payment status after %d attempts", retryCount)
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

func (a *App) signupHandler(w http.ResponseWriter, r *http.Request) {
	var user struct {
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the phone number is exactly 10 digits
	if len(user.PhoneNumber) != 10 {
		http.Error(w, "Phone number must be 10 digits", http.StatusBadRequest)
		return
	}

	// Insert the user into the database
	_, err := a.db.Exec("INSERT INTO users (phone_number) VALUES (?)", user.PhoneNumber)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User created successfully")
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	var user struct {
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Query the database to check if the user exists
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM users WHERE phone_number = ?", user.PhoneNumber).Scan(&count)
	if err != nil {
		http.Error(w, "Error checking user", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return success message on successful login
	fmt.Fprintln(w, "User logged in successfully")
}

// Handler function to get aggregated order data by phone number
func (a *App) getAggregatedOrdersByPhoneNumber(w http.ResponseWriter, r *http.Request) {
    // Get the phone number from the query parameters
    phoneNumber := r.URL.Query().Get("phoneNumber")

    if phoneNumber == "" {
        http.Error(w, "Phone number parameter is required", http.StatusBadRequest)
        return
    }

    // Execute the SQL query to get aggregated order data
    rows, err := a.db.Query(`
        SELECT 
            fund, 
            phone_number, 
            SUM(amount) AS total_amount, 
            SUM(units) AS total_units
        FROM 
            orders
        WHERE 
            phone_number = ? and status = "Succeeded"
        GROUP BY 
            fund, phone_number;
    `, phoneNumber)
    
    if err != nil {
        http.Error(w, "Error retrieving aggregated order data", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    // Create a map to store the aggregated order data in the desired format
    aggregatedOrdersMap := make(map[string]map[string]float64)

    // Iterate through the rows and aggregate the data
    for rows.Next() {
        var fundName, phoneNumber string
        var totalAmount, totalUnits float64
        err := rows.Scan(&fundName, &phoneNumber, &totalAmount, &totalUnits)
        if err != nil {
            http.Error(w, "Error parsing aggregated order data", http.StatusInternalServerError)
            return
        }

        // Fetch the fund market value from the cache
        fundC.Lock()
        fund, ok := fundC.Funds[fundName]
        fundC.Unlock()
        if !ok {
            http.Error(w, "Fund not found", http.StatusBadRequest)
            return
        }

        
		// Calculate the market value
		fundMarketValue := totalUnits * fund.MarketValue



        // Create a map for the current fund
        fundData := make(map[string]float64)
        fundData["total_amount"] = totalAmount
        fundData["market_value"] = fundMarketValue

        // Store the fund data in the aggregatedOrdersMap
        aggregatedOrdersMap[fundName] = fundData
    }

    // Convert aggregated order data map to JSON
    jsonAggregatedOrders, err := json.Marshal(aggregatedOrdersMap)
    if err != nil {
        http.Error(w, "Error encoding aggregated order data", http.StatusInternalServerError)
        return
    }

    // Set response headers and write JSON response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonAggregatedOrders)
}

// Fund represents a fund in an investment strategy.
type Funds struct {
    Name       string `json:"name"`
    Percentage int    `json:"percentage"`
}

// Strategy represents an investment strategy.
type Strategy struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Funds       []Funds `json:"funds"`
}

// ConvertStrategyJsonIntoMap converts a JSON representation of investment strategies
// into a map, where the key is the strategy name and the value is a slice of funds.
func convertStrategyJsonIntoMap() map[string][]Funds {
    // JSON data
    jsonData := `[
        {
            "name": "Arbitrage Strategy",
            "description": "This strategy is based on the concept of arbitrage. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms. This strategy is based on the concept of arbitrage. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms.",
            "funds": [
                {
                    "name": "Arbitrage Fund 1",
                    "percentage": 10
                },
                {
                    "name": "Arbitrage Fund 2",
                    "percentage": 20
                },
                {
                    "name": "Arbitrage Fund 3",
                    "percentage": 30
                },
                {
                    "name": "Arbitrage Fund 4",
                    "percentage": 40
                }
            ]
        },
		{
			"name": "Balanced Strategy",
			"description": "This strategy is based on the concept of balanced portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms. This strategy is based on the concept of balanced portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms.",
			"funds": [
				{
					"name": "Balanced Fund 1",
					"percentage": 20
				},
				{
					"name": "Balanced Fund 2",
					"percentage": 20
				},
				{
					"name": "Balanced Fund 3",
					"percentage": 5
				},
				{
					"name": "Balanced Fund 4",
					"percentage": 40
				},
				{
					"name": "Balanced Fund 5",
					"percentage": 15
				}
			]
		},
		{
			"name": "Growth Strategy",
			"description": "This strategy is based on the concept of growth portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms. This strategy is based on the concept of growth portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms.",
			"funds": [
				{
					"name": "Growth Fund 1",
					"percentage": 50
				},
				{
					"name": "Growth Fund 2",
					"percentage": 10
				},
				{
					"name": "Growth Fund 3",
					"percentage": 10
				},
				{
					"name": "Growth Fund 4",
					"percentage": 15
				}, 
				{
					"name": "Growth Fund 5",
					"percentage": 15
				}
			]
		}
    ]`

    // Unmarshal JSON
    var strategies []Strategy
    err := json.Unmarshal([]byte(jsonData), &strategies)
    if err != nil {
        fmt.Println("Error:", err)
        return nil
    }

    // Create a map
    strategyFundsMap := make(map[string][]Funds)
    for _, strategy := range strategies {
        strategyFundsMap[strategy.Name] = strategy.Funds
    }

    return strategyFundsMap
}