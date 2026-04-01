package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPPaymentClient implements domain.PaymentClient using HTTP.
type HTTPPaymentClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPPaymentClient creates a new HTTP payment client.
// httpClient should have a timeout configured (max 2 seconds as per requirements).
func NewHTTPPaymentClient(baseURL string, httpClient *http.Client) *HTTPPaymentClient {
	return &HTTPPaymentClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// AuthorizePaymentRequest is the request body for Payment Service.
type AuthorizePaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

// AuthorizePaymentResponse is the response from Payment Service.
type AuthorizePaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// AuthorizePayment calls the Payment Service to authorize a payment.
func (c *HTTPPaymentClient) AuthorizePayment(orderID string, amount int64) (transactionID string, status string, err error) {
	req := AuthorizePaymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/payments", c.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to call payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	var paymentResp AuthorizePaymentResponse
	err = json.Unmarshal(respBody, &paymentResp)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return paymentResp.TransactionID, paymentResp.Status, nil
}

// GetPaymentStatus retrieves the payment status from the Payment Service.
func (c *HTTPPaymentClient) GetPaymentStatus(orderID string) (status string, err error) {
	url := fmt.Sprintf("%s/payments/%s", c.baseURL, orderID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to call payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var paymentResp AuthorizePaymentResponse
	err = json.Unmarshal(respBody, &paymentResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return paymentResp.Status, nil
}
