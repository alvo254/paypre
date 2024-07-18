package mpesa

import (
	"bytes" // Add this import
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	ConsumerKey       string
	ConsumerSecret    string
	PassKey           string
	BusinessShortCode string
	Environment       string // "sandbox" or "production"
}

type MPesa struct {
	config Config
	client *http.Client
}

type STKPushResponse struct {
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	MerchantRequestID   string `json:"MerchantRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

func NewMPesa(config Config) *MPesa {
	return &MPesa{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *MPesa) getAPIBaseURL() string {
	if m.config.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (m *MPesa) getAccessToken() (string, error) {
	url := m.getAPIBaseURL() + "/oauth/v1/generate?grant_type=client_credentials"
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(m.config.ConsumerKey, m.config.ConsumerSecret)
	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func (m *MPesa) InitiateSTKPush(phoneNumber string, amount int) (*STKPushResponse, error) {
	url := m.getAPIBaseURL() + "/mpesa/stkpush/v1/processrequest"
	timestamp := time.Now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(m.config.BusinessShortCode + m.config.PassKey + timestamp))

	requestBody := map[string]string{
		"BusinessShortCode": m.config.BusinessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            fmt.Sprintf("%d", amount),
		"PartyA":            phoneNumber,
		"PartyB":            m.config.BusinessShortCode,
		"PhoneNumber":       phoneNumber,
		"CallBackURL":       "https://5c19-102-216-154-4.ngrok-free.app/b2c/result",
		"AccountReference":  "TestAccount",
		"TransactionDesc":   "Test Transaction",
	}
	jsonBody, _ := json.Marshal(requestBody)
	token, err := m.getAccessToken()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody)) // Use bytes.NewBuffer
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response STKPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
