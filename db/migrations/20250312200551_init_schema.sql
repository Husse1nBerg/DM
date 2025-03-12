-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE subscription_plans (
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    title VARCHAR(150) NOT NULL,
    description TEXT DEFAULT '',
    options TEXT DEFAULT '',
    monthly_price FLOAT CHECK (monthly_price >= 0) NOT NULL,
    annually_price FLOAT CHECK (annually_price >= 0) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    number_of_users INT CHECK (number_of_users >= 0) DEFAULT 0,
    trial_period INT CHECK (trial_period >= 0) DEFAULT 0
);

CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    zip_code VARCHAR(10) NOT NULL,
    country VARCHAR(100) NOT NULL
);

CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(100) NOT NULL,
    website VARCHAR(255) DEFAULT '',
    email VARCHAR(255),
    phone VARCHAR(17) DEFAULT '',
    company_size CHAR(1) CHECK (company_size IN ('s', 'm', 'l', 'x')) DEFAULT 's',
    logo VARCHAR(255),
    plan_id INT REFERENCES subscription_plans(id) ON DELETE SET NULL,
    address INT REFERENCES addresses(id) ON DELETE SET NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    first_name VARCHAR(150) NOT NULL,
    last_name VARCHAR(150) NOT NULL,
    phone VARCHAR(17) DEFAULT '',
    role_id INT REFERENCES roles(id) ON DELETE SET NULL,
    last_password_change TIMESTAMP NULL,
    failed_login_attempts INTEGER DEFAULT 0,
    lockout_until TIMESTAMP NULL,
    company_id INT REFERENCES companies(id) ON DELETE CASCADE,
    avatar VARCHAR(255),
    is_enterprise_user BOOLEAN DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE companies;
DROP TABLE roles;
DROP TABLE subscription_plans;
DROP TABLE addresses;
-- +goose StatementEnd
