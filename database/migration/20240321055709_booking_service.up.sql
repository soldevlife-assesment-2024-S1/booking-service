CREATE TABLE IF NOT EXISTS bookings (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id INT,
    ticket_detail_id INT,
    total_tickets INT,
    full_name TEXT,
    personal_id TEXT,
    booking_date TIMESTAMP,
    booking_expiration TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS payments (
    id BIGINT PRIMARY KEY,
    booking_id UUID REFERENCES booking(id),
    amount FLOAT,
    currency TEXT,
    status TEXT,
    payment_method TEXT,
    payment_date TIMESTAMP,
    payment_expiration TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);