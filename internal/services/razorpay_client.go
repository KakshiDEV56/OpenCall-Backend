package services

import (
	"errors"

	"github.com/razorpay/razorpay-go"
)

type RazorpayClient struct {
	client *razorpay.Client
}

func NewRazorpayClient(key, secret string) *RazorpayClient {
	return &RazorpayClient{
		client: razorpay.NewClient(key, secret),
	}
}

func (r *RazorpayClient) CreateOrder(amount int64, receipt string) (string, error) {
	data := map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  receipt,
	}

	body, err := r.client.Order.Create(data, nil)
	if err != nil {
		return "", err
	}

	orderID, ok := body["id"].(string)
	if !ok {
		return "", errors.New("invalid razorpay order response")
	}

	return orderID, nil
}
