package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// TokenRequest represents credentials for authentication
type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse represents the authentication token response
type TokenResponse struct {
	Token string `json:"token"`
}

// CollectRequest represents the payment collection request to CamPay
type CollectRequest struct {
	Amount            string `json:"amount"`
	Currency          string `json:"currency"`
	From              string `json:"from"`
	Description       string `json:"description"`
	ExternalReference string `json:"external_reference"`
}

// CollectResponse represents CamPay's response when initiating payment
type CollectResponse struct {
	Reference         string `json:"reference"`
	ExternalReference string `json:"external_reference"`
	Status            string `json:"status"`
	Amount            int    `json:"amount"`
	Currency          string `json:"currency"`
	Operator          string `json:"operator"`
	Code              string `json:"code"`
	OperatorReference string `json:"operator_reference"`
}

// TransactionResponse represents the transaction status from CamPay
type TransactionResponse struct {
	Reference         string  `json:"reference"`
	ExternalReference string  `json:"external_reference"`
	Status            string  `json:"status"`
	Amount            float64 `json:"amount"` // CamPay returns this as number
	Currency          string  `json:"currency"`
	Operator          string  `json:"operator"`
	Code              string  `json:"code"`
	OperatorReference string  `json:"operator_reference"`
	Description       string  `json:"description"`
}

// ErrorResponse represents CamPay API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func main() {
	err := run()
	if err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Load environment variables from .env file
	if _, err := os.Stat(".env"); err == nil {
		err = godotenv.Load()
		if err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Get credentials from environment
	username := os.Getenv("APP_USERNAME")
	password := os.Getenv("APP_PASSWORD")
	environment := os.Getenv("ENVIRONMENT") 

	if username == "" || password == "" {
		return fmt.Errorf("APP_USERNAME and APP_PASSWORD must be set in .env file")
	}

	if environment == "" {
		environment = "DEV" // Default to development/demo
	}

	// Determine API base URL based on environment
	var apiBaseURL string
	if environment == "PROD" {
		apiBaseURL = "https://www.campay.net/api"
	} else {
		apiBaseURL = "https://demo.campay.net/api"
	}

	fmt.Println("===  CamPay Mobile Money Payment System ===")
	fmt.Printf("Environment: %s\n\n", environment)

	// Step 1: Get authentication token
	fmt.Println(" Authenticating with CamPay...")
	token, err := getAuthToken(apiBaseURL, username, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println("✓ Authentication successful!")

	// Step 2: Get user input
	phoneNumber, err := promptUser(" Enter mobile money number (e.g., 237670123456): ")
	if err != nil {
		return err
	}

	// Validate phone number format
	if !strings.HasPrefix(phoneNumber, "237") || len(phoneNumber) != 12 {
		return fmt.Errorf("invalid phone number format. Must start with 237 and be 12 digits total")
	}

	amountStr, err := promptUser(" Enter amount (XAF, no decimals): ")
	if err != nil {
		return err
	}

	// Validate amount is a positive integer
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		return fmt.Errorf("invalid amount. Must be a positive integer")
	}

	description, err := promptUser(" Enter description: ")
	if err != nil {
		return err
	}

	// Generate a unique external reference for this transaction
	externalRef := fmt.Sprintf("TXN-%d", time.Now().Unix())

	// Step 3: Create payment collection request
	collectReq := CollectRequest{
		Amount:            amountStr,
		Currency:          "XAF",
		From:              phoneNumber,
		Description:       description,
		ExternalReference: externalRef,
	}

	fmt.Println("\n Initiating payment collection...")

	// Step 4: Send collect request to CamPay
	reference, err := collectPayment(apiBaseURL, token, collectReq)
	if err != nil {
		return fmt.Errorf("failed to initiate payment: %w", err)
	}

	fmt.Printf("\n✓ Payment collection initiated successfully!")
	fmt.Printf("\nTransaction Reference: %s\n", reference)
	fmt.Printf("External Reference: %s\n", externalRef)
	fmt.Println("\n Please check your phone for USSD prompt...")
	fmt.Println(" Enter your mobile money PIN to complete the payment...")
	fmt.Println("\nWaiting for transaction completion...")

	// Step 5: Poll for transaction status
	finalStatus, err := pollTransactionStatus(apiBaseURL, token, reference)
	if err != nil {
		return fmt.Errorf("failed to get transaction status: %w", err)
	}

	// Step 6: Display final status
	displayFinalStatus(finalStatus)

	return nil
}

// getAuthToken authenticates with CamPay and returns access token
func getAuthToken(baseURL, username, password string) (string, error) {
	tokenReq := TokenRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	url := baseURL + "/token/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(body, &errResp)
		return "", fmt.Errorf("authentication failed (status %d): %s - %s", resp.StatusCode, errResp.Code, errResp.Message)
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return tokenResp.Token, nil
}

// promptUser displays a prompt and reads user input
func promptUser(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(input), nil
}

// collectPayment initiates a payment collection from user's mobile money
func collectPayment(baseURL, token string, collect CollectRequest) (string, error) {
	jsonData, err := json.Marshal(collect)
	if err != nil {
		return "", fmt.Errorf("failed to marshal collect request: %w", err)
	}

	url := baseURL + "/collect/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// CamPay returns 200 for successful collect initiation
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(body, &errResp)
		return "", fmt.Errorf("collect request failed (status %d): %s - %s", resp.StatusCode, errResp.Code, errResp.Message)
	}

	var collectResp CollectResponse
	err = json.Unmarshal(body, &collectResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse collect response: %w", err)
	}

	return collectResp.Reference, nil
}

// pollTransactionStatus continuously checks transaction status until complete
func pollTransactionStatus(baseURL, token, reference string) (*TransactionResponse, error) {
	maxAttempts := 60 // 5 minutes (60 * 5 seconds)
	interval := 5 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		status, err := checkTransactionStatus(baseURL, token, reference)
		if err != nil {
			return nil, err
		}

		// Check if transaction is in a final state
		switch strings.ToUpper(status.Status) {
		case "SUCCESSFUL":
			return status, nil
		case "FAILED":
			return status, nil
		case "PENDING":
			fmt.Printf(" Status: PENDING (attempt %d/%d, checking again in %v...)\n", attempt+1, maxAttempts, interval)
		default:
			fmt.Printf(" Status: %s (checking again in %v...)\n", status.Status, interval)
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("transaction status check timed out after %d attempts", maxAttempts)
}

// checkTransactionStatus makes a single API call to check transaction status
func checkTransactionStatus(baseURL, token, reference string) (*TransactionResponse, error) {
	url := fmt.Sprintf("%s/transaction/%s/", baseURL, reference)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(body, &errResp)
		return nil, fmt.Errorf("status check failed (status %d): %s - %s", resp.StatusCode, errResp.Code, errResp.Message)
	}

	var txnResp TransactionResponse
	err = json.Unmarshal(body, &txnResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return &txnResp, nil
}

// displayFinalStatus prints the final transaction status in a formatted way
func displayFinalStatus(status *TransactionResponse) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("           TRANSACTION FINAL STATUS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Reference:           %s\n", status.Reference)
	fmt.Printf("External Reference:  %s\n", status.ExternalReference)
	fmt.Printf("Status:              %s\n", status.Status)
	fmt.Printf("Amount:              %.0f %s\n", status.Amount, status.Currency)
	fmt.Printf("Operator:            %s\n", status.Operator)
	fmt.Printf("Description:         %s\n", status.Description)
	fmt.Printf("Transaction Code:    %s\n", status.Code)
	fmt.Printf("Operator Reference:  %s\n", status.OperatorReference)
	fmt.Println(strings.Repeat("=", 60))

	switch strings.ToUpper(status.Status) {
	case "SUCCESSFUL":
		fmt.Println("Payment completed successfully!")
		fmt.Println("\n Thank you for using CamPay!")
	case "FAILED":
		fmt.Println("❌ Payment failed or was cancelled")
		fmt.Println("\n Please try again or contact support if the issue persists")
	default:
		fmt.Printf("  Payment status: %s\n", status.Status)
	}
}