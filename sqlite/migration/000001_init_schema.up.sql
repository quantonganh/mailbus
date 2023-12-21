CREATE TABLE subscriptions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    status        TEXT NOT NULL,
    subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subscription_tokens (
    subscription_token TEXT NOT NULL,
    subscriber_id      INTEGER NOT NULL REFERENCES subscriptions (id),

    PRIMARY KEY (subscription_token)
);
