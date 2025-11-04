CREATE TABLE IF NOT EXISTS subscriptions (
    service_name VARCHAR NOT NULL,
    price INTEGER NOT NULL,
    user_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    start_date VARCHAR NOT NULL
);