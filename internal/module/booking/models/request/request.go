package request

type BookTicket struct {
	TicketDetailID int64  `json:"ticket_detail_id" validate:"required"`
	FullName       string `json:"full_name" validate:"required"`
	PersonalID     string `json:"personal_id" validate:"required"`
}

type Payment struct {
	BookingID    int64   `json:"booking_id" validate:"required"`
	TotalAmount  float64 `json:"total_amount" validate:"required"`
	PaymetMethod string  `json:"payment_method" validate:"required"`
}
