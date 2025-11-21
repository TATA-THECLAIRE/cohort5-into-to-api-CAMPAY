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

/* ============================================================
   ===============  REQUEST / RESPONSE MODELS  =================
   ============================================================ */

type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type CollectRequest struct {
	Amount            int    `json:"amount"`
	Currency          string `json:"currency"`
	From              string `json:"from"`
	Description       string `json:"description"`
	ExternalReference string `json:"external_reference"`
}

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

type TransactionResponse struct {
	Reference         string  `json:"reference"`
	ExternalReference string  `json:"external_reference"`
	Status            string  `json:"status"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Operator          string  `json:"operator"`
	Code              string  `json:"code"`
	OperatorReference string  `json:"operator_reference"`
	Description       string  `json:"description"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

/* ============================================================
   ========================= MAIN ==============================
   ============================================================ */

func main() {
	if err := run(); err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Load .env values
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("failed to load .env: %w", err)
		}
	}

	username := os.Getenv("APP_USERNAME")
	password := os.Getenv("APP_PASSWORD")
	env := os.Getenv("ENVIRONMENT")

	if username == "" || password == "" {
		return fmt.Errorf("APP_USERNAME and APP_PASSWORD must be set")
	}
	if env == "" {
		env = "DEV"
	}

	apiBaseURL := map[bool]string{
		true:  "https://www.campay.net/api",
		false: "https://demo.campay.net/api",
	}[env == "PROD"]

	fmt.Println("=== CamPay Mobile Money Payment System ===")
	fmt.Printf("Environment: %s\n\n", env)

	// Authenticate
	fmt.Println("🔐 Authenticating...")
	token, err := getAuthToken(apiBaseURL, username, password)
	if err != nil {
		return err
	}
	fmt.Println("✓ Authentication successful")

	// User Input
	phone, err := promptPhone()
	if err != nil {
		return err
	}

	amount, err := promptAmount()
	if err != nil {
		return err
	}

	description, err := promptUser("Enter description: ")
	if err != nil {
		return err
	}

	externalRef := fmt.Sprintf("TXN-%d", time.Now().Unix())

	collectReq := CollectRequest{
		Amount:            amount,
		Currency:          "XAF",
		From:              phone,
		Description:       description,
		ExternalReference: externalRef,
	}

	fmt.Println("\n📲 Initiating payment...")

	// Collect request
	reference, err := collectPayment(apiBaseURL, token, collectReq)
	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Payment initiated\nReference: %s\n", reference)
	fmt.Println("Please check your phone for USSD popup...")

	// Wait for status
	finalStatus, err := pollTransactionStatus(apiBaseURL, token, reference)
	if err != nil {
		return err
	}

	displayFinalStatus(finalStatus)
	return nil
}

/* ============================================================
   ====================== HELPER FUNCTIONS =====================
   ============================================================ */

var httpClient = &http.Client{Timeout: 30 * time.Second}

// =============================================================
// Authentication
// =============================================================

func getAuthToken(baseURL, username, password string) (string, error) {
	reqBody, _ := json.Marshal(TokenRequest{Username: username, Password: password})
	req, err := http.NewRequest("POST", baseURL+"/token/", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", formatAPIError(resp.StatusCode, body)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

// =============================================================
// User Input
// =============================================================

func promptUser(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return promptUser(prompt)
	}
	return input, nil
}

func promptPhone() (string, error) {
	phone, err := promptUser("Enter mobile money number (e.g., 670123456 or 237670123456): ")
	if err != nil {
		return "", err
	}

	// Normalize
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")

	// Local number
	if len(phone) == 9 && phone[0] == '6' {
		phone = "237" + phone
	}

	if !strings.HasPrefix(phone, "237") || len(phone) != 12 {
		return "", fmt.Errorf("invalid phone number format")
	}
	return phone, nil
}

func promptAmount() (int, error) {
	amtStr, err := promptUser("Enter amount (XAF): ")
	if err != nil {
		return 0, err
	}

	amount, err := strconv.Atoi(amtStr)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("amount must be a positive integer")
	}

	return amount, nil
}

// =============================================================
// Payment Collect
// =============================================================

func collectPayment(baseURL, token string, collect CollectRequest) (string, error) {
	reqBody, _ := json.Marshal(collect)

	req, err := http.NewRequest("POST", baseURL+"/collect/", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", formatAPIError(resp.StatusCode, body)
	}

	var collectResp CollectResponse
	if err := json.Unmarshal(body, &collectResp); err != nil {
		return "", err
	}

	return collectResp.Reference, nil
}

// =============================================================
// Poll for Status
// =============================================================

func pollTransactionStatus(baseURL, token, reference string) (*TransactionResponse, error) {
	const maxAttempts = 40
	const interval = 5 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		status, err := checkTransactionStatus(baseURL, token, reference)
		if err != nil {
			return nil, err
		}

		s := normalizeStatus(status.Status)

		if s == "SUCCESSFUL" || s == "FAILED" {
			return status, nil
		}

		fmt.Printf("Status: %s (attempt %d/%d)\n", s, attempt, maxAttempts)
		time.Sleep(interval)
	}

	return nil, fmt.Errorf("transaction polling timed out")
}

func checkTransactionStatus(baseURL, token, reference string) (*TransactionResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/transaction/%s/", baseURL, reference), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, formatAPIError(resp.StatusCode, body)
	}

	var txn TransactionResponse
	if err := json.Unmarshal(body, &txn); err != nil {
		return nil, err
	}

	return &txn, nil
}

// =============================================================
// Helpers
// =============================================================

func normalizeStatus(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func formatAPIError(status int, body []byte) error {
	var er ErrorResponse
	if json.Unmarshal(body, &er) == nil && er.Message != "" {
		return fmt.Errorf("API error (%d): %s - %s", status, er.Code, er.Message)
	}
	return fmt.Errorf("API error (%d): %s", status, string(body))
}

// =============================================================
// Display Result
// =============================================================

func displayFinalStatus(s *TransactionResponse) {
	fmt.Println("\n============================================================")
	fmt.Println("                 TRANSACTION FINAL STATUS")
	fmt.Println("============================================================")

	fmt.Printf("Reference:           %s\n", s.Reference)
	fmt.Printf("External Reference:  %s\n", s.ExternalReference)
	fmt.Printf("Status:              %s\n", s.Status)
	fmt.Printf("Amount:              %.0f %s\n", s.Amount, s.Currency)
	fmt.Printf("Operator:            %s\n", s.Operator)
	fmt.Printf("Description:         %s\n", s.Description)
	fmt.Printf("Code:                %s\n", s.Code)
	fmt.Printf("Operator Reference:  %s\n", s.OperatorReference)
	fmt.Println("============================================================")

	switch normalizeStatus(s.Status) {
	case "SUCCESSFUL":
		fmt.Println("🎉 Payment successful!")
	case "FAILED":
		fmt.Println("❌ Payment failed")
	default:
		fmt.Println("⚠ Unknown status:", s.Status)
	}
}
