CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50),
    age INT,
    email TEXT UNIQUE,
    password TEXT,
    role user_role DEFAULT 'user'
);