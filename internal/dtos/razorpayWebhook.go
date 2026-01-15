package dtos

type RazorpayWebhookEvent struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID      string `json:"id"`
				OrderID string `json:"order_id"`
				Status  string `json:"status"`
				Amount  int64  `json:"amount"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}
