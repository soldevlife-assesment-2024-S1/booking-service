package response

type UserServiceValidate struct {
	IsValid bool `json:"is_valid"`
	UserID  int  `json:"user_id"`
}

type BookedTicket struct {
	ID            string  `json:"id"`
	FullName      string  `json:"full_name"`
	PersonalID    string  `json:"personal_id"`
	BookingDate   string  `json:"booking_date"`
	PaymentExpiry string  `json:"payment_expiry"`
	TotalAmount   float64 `json:"total_amount"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
}
