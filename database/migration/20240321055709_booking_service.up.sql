CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS bookings (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id INT,
    ticket_detail_id INT,
    total_tickets INT,
    full_name TEXT,
    personal_id TEXT,
    booking_date TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);

