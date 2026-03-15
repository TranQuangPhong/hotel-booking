CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(200) NOT NULL,
    role VARCHAR(20) DEFAULT 'USER',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users (username, email, password, role)
VALUES ('demo_user', 'demo@example.com', 'hashed_password', 'USER')
ON CONFLICT (username) DO NOTHING;


CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    event_id VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO transactions (user_id, event_id, amount, currency, status)
VALUES (1, 'evt_demo', 100.00, 'USD', 'SUCCESS')
ON CONFLICT DO NOTHING;
