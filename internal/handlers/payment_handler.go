package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
)

type PaymentHandler struct {
	service *services.PaymentService
}

func NewPaymentHandler(s *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: s}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dtos.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	payment, err := h.service.CreatePayment(c.Request.Context(), req.BookingID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_id":        payment.ID,
		"razorpay_order_id": payment.GatewayOrderID,
		"amount":            payment.Amount,
		"currency":          payment.Currency,
	})
}

func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	var req dtos.VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.VerifyPayment(
		c.Request.Context(),
		req.PaymentID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "payment verified"})
}

func (h *PaymentHandler) RazorpayWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Razorpay-Signature")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.service.VerifyWebhookSignature(body, signature); err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	var event dtos.RazorpayWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	switch event.Event {
	case "payment.captured":
		if err := h.service.HandlePaymentCaptured(event); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	case "payment.failed":
		if err := h.service.HandlePaymentFailed(event); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Status(http.StatusOK)
}
