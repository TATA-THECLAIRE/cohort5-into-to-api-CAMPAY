# CamPay Mobile Money Payment System

A Go-based command-line application for processing mobile money payments in Cameroon using the CamPay API. Supports MTN Mobile Money and Orange Money.


## Overview

This application integrates with CamPay, a payment gateway that enables merchants to accept and process payments via MTN Mobile Money and Orange Mobile Money in Cameroon. Users can:

1. Enter their mobile money phone number (MTN or Orange)
2. Specify a payment amount in XAF (Central African Francs)
3. Add a transaction description
4. Receive USSD prompt on their phone
5. Complete payment with their PIN
6. View the final transaction status

##  Features

- **Secure Authentication**: Uses token-based authentication with username/password credentials
- **Dual Environment Support**: Test on demo site or run live transactions
- **Real-time Status Polling**: Automatically checks transaction status every 5 seconds
- **MTN & Orange Support**: Works with both major mobile money providers in Cameroon
- **Input Validation**: Validates phone numbers and amounts before API calls
- **Comprehensive Error Handling**: Clear error messages with CamPay error codes
- **User-Friendly CLI**: Clean interface with emojis and formatted output

##  Prerequisites

- **Go 1.21 or higher**: [Download Go](https://golang.org/dl/)
- **Git**: For version control
- **CamPay Account**: Free registration at https://demo.campay.net or https://www.campay.net
- **Mobile Money Account**: MTN or Orange Money account for testing (Cameroon)

## CamPay Account Setup

### Step 1: Register on CamPay

**For Testing (Recommended for Development):**
1. Go to https://demo.campay.net
2. Click "Sign Up" and create an account
3. Verify your email address
4. Log in to your dashboard

**For Live Transactions:**
1. Go to https://www.campay.net
2. Register for a live account
3. Complete KYC verification

### Step 2: Create an Application

1. Once logged in, navigate to "Applications" or "My Apps"
2. Click "Register Application" or "Create New App"
3. Fill in application details:
   - **Name**: e.g., "My Payment App"
   - **Description**: Brief description of your app
4. Submit the application

### Step 3: Get API Credentials

1. After creating the application, expand/click on it
2. You'll see **APP KEYS** section with:
   - **Username**: Your APP_USERNAME
   - **Password**: Your APP_PASSWORD
   - **Token**: (optional, for permanent token method)
3. Copy these credentials - you'll need them for `.env` file

### Important Notes:
- Demo environment credentials only work on demo.campay.net
- Production credentials only work on www.campay.net
- Phone numbers must start with country code 237
- Only MTN and Orange phone numbers are currently supported

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/campay-mobile-money.git
cd campay-mobile-money
```

### 2. Install Dependencies

```bash
go mod download
```

This installs the `godotenv` package for environment variable management.

### 3. Set Up Environment Variables

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your CamPay credentials
nano .env  # or use your preferred editor
```

##  Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
APP_USERNAME="your-campay-app-username"
APP_PASSWORD="your-campay-app-password"
ENVIRONMENT="DEV"
```

**Configuration Options:**

| Variable | Description | Values | Required |
|----------|-------------|--------|----------|
| `APP_USERNAME` | Your CamPay application username | From CamPay dashboard | Yes |
| `APP_PASSWORD` | Your CamPay application password | From CamPay dashboard | Yes |
| `ENVIRONMENT` | Operating environment | `DEV` or `PROD` | No (defaults to DEV) |

**Important Security Notes:**
-  Never commit `.env` to Git (already in `.gitignore`)
-  Keep credentials secret
-  Use demo environment for testing
-  Different credentials for DEV vs PROD

## Usage

### Running the Application

```bash
go run main.go
```

### Example Session

```
===  CamPay Mobile Money Payment System ===
Environment: DEV

Authenticating with CamPay...
✓ Authentication successful!

 Enter mobile money number (e.g., 237670123456): 237670123456
 Enter amount (XAF, no decimals): 1000
 Enter description: Test payment for groceries

 Initiating payment collection...

✓ Payment collection initiated successfully!
Transaction Reference: bcedde9b-62a7-4421-96ac-2e6179552a1a
External Reference: TXN-1700000000

 Please check your phone for USSD prompt...
 Enter your mobile money PIN to complete the payment...

Waiting for transaction completion...

 Status: PENDING (attempt 1/60, checking again in 5s...)
Status: PENDING (attempt 2/60, checking again in 5s...)

============================================================
            TRANSACTION FINAL STATUS
============================================================
Reference:           bcedde9b-62a7-4421-96ac-2e6179552a1a
External Reference:  TXN-1700000000
Status:              SUCCESSFUL
Amount:              1000 XAF
Operator:            MTN
Description:         Test payment for groceries
Transaction Code:    CP201027U00005
Operator Reference:  1880106956
============================================================
 Payment completed successfully!

Thank you for using CamPay!
```

### Phone Number Format

Phone numbers must start with the Cameroon country code 237:

-  Correct: `237670123456` (12 digits total)
-  Correct: `237690123456` (Orange)
-  Wrong: `670123456` (missing country code)
-  Wrong: `+237670123456` (no + symbol)

### Amount Format

Amounts must be integers without decimals:

-  Correct: `1000`
-  Correct: `500`
-  Wrong: `100.50` (no decimals allowed)
- Wrong: `-100` (must be positive)

### Building the Executable

```bash
# Build for your current platform
go build -o campay-payment

# Run the built executable
./campay-payment
```

### Cross-Platform Building

```bash
# For Windows
GOOS=windows GOARCH=amd64 go build -o campay-payment.exe

# For Linux
GOOS=linux GOARCH=amd64 go build -o campay-payment

# For macOS
GOOS=darwin GOARCH=amd64 go build -o campay-payment
```

## 🔍 Code Explanation

### Project Structure

```
campay-mobile-money/
├── main.go           # Main application code
├── go.mod            # Go module dependencies
├── go.sum            # Dependency checksums
├── .env              # Environment variables (not in Git)
├── .env.example      # Environment template
├── .gitignore        # Git ignore rules
└── README.md         # This file
```

### Key Components

#### 1. Data Structures

```go
// Authentication
type TokenRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type TokenResponse struct {
    Token string `json:"token"`
}

// Payment Collection
type CollectRequest struct {
    Amount            string `json:"amount"`
    Currency          string `json:"currency"`
    From              string `json:"from"`              // Phone number
    Description       string `json:"description"`
    ExternalReference string `json:"external_reference"` // Your tracking ID
}

type CollectResponse struct {
    Reference         string `json:"reference"`          // CamPay transaction ID
    Status            string `json:"status"`
    Amount            int    `json:"amount"`
    Currency          string `json:"currency"`
    Operator          string `json:"operator"`          // MTN or Orange
    Code              string `json:"code"`              // Transaction code
}

// Transaction Status
type TransactionResponse struct {
    Reference         string `json:"reference"`
    ExternalReference string `json:"external_reference"`
    Status            string `json:"status"`
    Amount            string `json:"amount"`
    Currency          string `json:"currency"`
    Operator          string `json:"operator"`
    Description       string `json:"description"`
}
```

#### 2. Authentication Flow

The application uses token-based authentication where you first obtain a temporary access token using username and password:

```go
func getAuthToken(baseURL, username, password string) (string, error) {
    // 1. Create token request with credentials
    tokenReq := TokenRequest{
        Username: username,
        Password: password,
    }
    
    // 2. POST to /token/ endpoint
    url := baseURL + "/token/"
    
    // 3. Receive token in response
    var tokenResp TokenResponse
    json.Unmarshal(body, &tokenResp)
    
    return tokenResp.Token, nil
}
```

#### 3. Payment Collection Flow

```go
func collectPayment(baseURL, token string, collect CollectRequest) (string, error) {
    // 1. Prepare collect request with payment details
    // 2. POST to /collect/ endpoint with Token authentication
    req.Header.Set("Authorization", "Token "+token)
    
    // 3. Get reference from response
    var collectResp CollectResponse
    json.Unmarshal(body, &collectResp)
    
    return collectResp.Reference, nil
}
```

**What happens on user's phone:**
1. CamPay sends USSD prompt to the phone number
2. User sees payment request with amount and description
3. User enters their mobile money PIN
4. Transaction is processed by MTN/Orange
5. CamPay receives confirmation from operator

#### 4. Status Polling

```go
func pollTransactionStatus(baseURL, token, reference string) (*TransactionResponse, error) {
    for attempt := 0; attempt < 60; attempt++ {
        status, _ := checkTransactionStatus(baseURL, token, reference)
        
        switch status.Status {
        case "SUCCESSFUL":
            return status, nil  // Payment complete
        case "FAILED":
            return status, nil  // Payment failed
        case "PENDING":
            time.Sleep(5 * time.Second)  // Wait and retry
        }
    }
}
```

#### 5. Transaction Status Check

```go
func checkTransactionStatus(baseURL, token, reference string) (*TransactionResponse, error) {
    // GET /transaction/{reference}/
    url := fmt.Sprintf("%s/transaction/%s/", baseURL, reference)
    req.Header.Set("Authorization", "Token "+token)
    
    var txnResp TransactionResponse
    json.Unmarshal(body, &txnResp)
    
    return &txnResp, nil
}
```

### Important Go Patterns Used

#### Input Validation

```go
// Validate phone number
if !strings.HasPrefix(phoneNumber, "237") || len(phoneNumber) != 12 {
    return fmt.Errorf("invalid phone number format")
}

// Validate amount
amount, err := strconv.Atoi(amountStr)
if err != nil || amount <= 0 {
    return fmt.Errorf("invalid amount")
}
```

#### Error Handling with CamPay Error Codes

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

if resp.StatusCode != http.StatusOK {
    var errResp ErrorResponse
    json.Unmarshal(body, &errResp)
    return "", fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
}
```

## 🔌 CamPay API Integration

### API Endpoints

#### Base URLs
- **Demo/Development**: `https://demo.campay.net/api`
- **Production**: `https://www.campay.net/api`

#### 1. Get Authentication Token

**Endpoint**: `POST /token/`

**Request**:
```json
{
  "username": "your-app-username",
  "password": "your-app-password"
}
```

**Response**:
```json
{
  "token": "abc123xyz456token789"
}
```

#### 2. Collect Payment

**Endpoint**: `POST /collect/`

**Headers**:
```
Content-Type: application/json
Authorization: Token abc123xyz456token789
```

**Request**:
```json
{
  "amount": "1000",
  "currency": "XAF",
  "from": "237670123456",
  "description": "Payment for services",
  "external_reference": "ORDER-12345"
}
```

**Response**:
```json
{
  "reference": "bcedde9b-62a7-4421-96ac-2e6179552a1a",
  "external_reference": "ORDER-12345",
  "status": "PENDING",
  "amount": 1000,
  "currency": "XAF",
  "operator": "MTN",
  "code": "CP201027U00005",
  "operator_reference": ""
}
```

#### 3. Check Transaction Status

**Endpoint**: `GET /transaction/{reference}/`

**Headers**:
```
Authorization: Token abc123xyz456token789
```

**Response**:
```json
{
  "reference": "bcedde9b-62a7-4421-96ac-2e6179552a1a",
  "external_reference": "ORDER-12345",
  "status": "SUCCESSFUL",
  "amount": "1000",
  "currency": "XAF",
  "operator": "MTN",
  "code": "CP201027U00005",
  "operator_reference": "1880106956",
  "description": "Payment for services"
}
```

### Transaction Status Values

| Status | Description |
|--------|-------------|
| `PENDING` | Payment initiated, waiting for user confirmation |
| `SUCCESSFUL` | Payment completed successfully |
| `FAILED` | Payment failed or was cancelled by user |

## Error Codes

CamPay API returns specific error codes for different failure scenarios:

| Error Code | Description | Solution |
|------------|-------------|----------|
| ER101 | Invalid phone number format | Ensure number starts with 237 (e.g., 237670123456) |
| ER102 | Unsupported carrier | Only MTN (237 67/65/68) and Orange (237 69) are supported |
| ER201 | Invalid amount | Amount must be an integer, no decimals allowed |
| ER301 | Insufficient balance | User doesn't have enough money in their mobile money account |

## Troubleshooting

### Common Issues

#### 1. "APP_USERNAME and APP_PASSWORD must be set"

**Problem**: `.env` file not found or credentials not set

**Solution**:
```bash
cp .env.example .env
# Edit .env and add your CamPay credentials
```

#### 2. "authentication failed"

**Problem**: Invalid credentials or wrong environment

**Solutions**:
- Verify username/password from CamPay dashboard
- Ensure you're using demo credentials with demo environment
- Check for typos in `.env` file
- Make sure no extra spaces around `=` sign

#### 3. "invalid phone number format"

**Problem**: Phone number doesn't match required format

**Solution**:
- Must start with `237` (Cameroon country code)
- Must be exactly 12 digits
- Examples: `237670123456` (MTN), `237690123456` (Orange)

#### 4. "invalid amount"

**Problem**: Amount contains decimals or is not a number

**Solution**:
- Use integers only: `1000`, not `1000.50`
- Amount must be positive
- No currency symbols

#### 5. "collect request failed: ER102"

**Problem**: Phone number carrier not supported

**Solution**:
- MTN numbers: 237 67XXXXXXX, 237 65XXXXXXX, 237 68XXXXXXX
- Orange numbers: 237 69XXXXXXX
- Camtel and other carriers are not supported

#### 6. "transaction status check timed out"

**Problem**: User didn't complete payment within 5 minutes

**Solutions**:
- Check if USSD prompt appeared on phone
- Verify sufficient balance in mobile money account
- Try with a smaller amount for testing
- Check phone network connection

### Debug Mode

To see raw API responses, add this after reading response body:

```go
fmt.Println("Debug - Response Status:", resp.StatusCode)
fmt.Println("Debug - Response Body:", string(body))
```

## 🧪 Testing

### Test Numbers (Demo Environment)

When using the demo environment (`ENVIRONMENT="DEV"`), you can use these test scenarios:

1. **Test with your real phone number**: The demo won't actually charge you
2. **Phone numbers**: Must be real Cameroon MTN/Orange numbers
3. **Amounts**: Use small amounts like 100 or 500 XAF for testing

### Testing Checklist

- [ ] Authentication works with demo credentials
- [ ] Phone number validation catches invalid formats
- [ ] Amount validation catches decimals and negatives
- [ ] USSD prompt appears on phone
- [ ] Status polling shows PENDING correctly
- [ ] Successful payment shows correct status
- [ ] Cancelled payment shows FAILED status
- [ ] Error messages are clear and helpful

