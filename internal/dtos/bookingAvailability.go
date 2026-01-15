package dtos

type AvailableSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type AvailabilityResponse struct {
	Date  string          `json:"date"`
	Slots []AvailableSlot `json:"slots"`
}
