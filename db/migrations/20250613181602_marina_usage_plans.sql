-- +goose Up
-- +goose StatementBegin
CREATE TABLE notes_messages_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    monthly_price NUMERIC(10, 2),
    text_limit TEXT DEFAULT 'Unlimited Texts',
    email_limit TEXT DEFAULT 'Unlimited Emails',
    user_limit TEXT DEFAULT 'Unlimited Users',
    is_most_popular BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE storage_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    monthly_price NUMERIC(10, 2),
    storage_limit_gb TEXT DEFAULT 'Unlimited Storage',
    user_limit TEXT DEFAULT 'Unlimited Users',
    is_most_popular BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

ALTER TABLE marinas ADD COLUMN notes_messages_plan_id UUID REFERENCES notes_messages_plans(id);
ALTER TABLE marinas ADD COLUMN storage_plan_id UUID REFERENCES storage_plans(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN notes_messages_plan_id;
ALTER TABLE marinas DROP COLUMN storage_plan_id;
DROP TABLE IF EXISTS notes_messages_plans;
DROP TABLE IF EXISTS storage_plans;
-- +goose StatementEnd
