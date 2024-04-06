package request

type BookTicket struct {
	TicketDetailID int64  `json:"ticket_detail_id" validate:"required"`
	FullName       string `json:"full_name" validate:"required"`
	PersonalID     string `json:"personal_id" validate:"required"`
	UserID         int64  `json:"user_id" validate:"required"`
	TotalTickets   int    `json:"total_tickets" validate:"required"`
	EmailRecipient string `json:"email_recipient"`
}

type Payment struct {
	BookingID    string  `json:"booking_id" validate:"required"`
	TotalAmount  float64 `json:"total_amount" validate:"required"`
	PaymetMethod string  `json:"payment_method" validate:"required"`
}

type PaymentCancellation struct {
	BookingID string `json:"booking_id" validate:"required"`
}

type PoisonedQueue struct {
	TopicTarget string      `json:"topic_target" validate:"required"`
	ErrorMsg    string      `json:"error_msg" validate:"required"`
	Payload     interface{} `json:"payload" validate:"required"`
}

type DecrementStockTicket struct {
	TicketDetailID int64 `json:"ticket_detail_id" validate:"required"`
	TotalTickets   int   `json:"total_tickets" validate:"required"`
}

type PaymentExpiration struct {
	BookingID      string `json:"booking_id" validate:"required"`
	TicketDetailID int64  `json:"ticket_detail_id" validate:"required"`
	TotalTickets   int    `json:"total_tickets" validate:"required"`
}

type NotificationMessage struct {
	Message        string `json:"message" validate:"required"`
	EmailRecipient string `json:"email_recipient" validate:"required"`
}

type NotificationInvoice struct {
	BookingID         string  `json:"booking_id" validate:"required"`
	PaymentAmount     float64 `json:"payment_amount" validate:"required"`
	PaymentExpiration string  `json:"payment_expiration" validate:"required"`
	EmailRecipient    string  `json:"email_recipient" validate:"required"`
}

type NotificationPayment struct {
	BookingID      string `json:"booking_id" validate:"required"`
	Message        string `json:"message" validate:"required"`
	PaymentMethod  string `json:"payment_method" validate:"required"`
	EmailRecipient string `json:"email_recipient" validate:"required"`
}
