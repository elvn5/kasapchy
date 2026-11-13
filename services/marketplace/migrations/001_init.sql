CREATE TABLE IF NOT EXISTS market_users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), email text NOT NULL UNIQUE,
 role text NOT NULL CHECK(role IN ('Buyer','Seller')), name text NOT NULL DEFAULT '',
 phone text NOT NULL DEFAULT '', provider text NOT NULL DEFAULT 'Email', created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS market_sessions (
 token_hash text PRIMARY KEY, user_id uuid NOT NULL REFERENCES market_users(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS market_challenges (
 id text PRIMARY KEY, email text NOT NULL UNIQUE, role text NOT NULL, code_hash text NOT NULL,
 attempts int NOT NULL DEFAULT 0, expires_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS market_limits (
 key text PRIMARY KEY, count int NOT NULL, until_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS market_products (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, seller_id uuid NOT NULL REFERENCES market_users(id) ON DELETE CASCADE,
 name text NOT NULL, category text NOT NULL CHECK(category IN ('Beef','Lamb','Poultry')),
 price int NOT NULL CHECK(price BETWEEN 1 AND 100000), stock int NOT NULL CHECK(stock>=0),
 description text NOT NULL DEFAULT '', created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS market_orders (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 buyer_id uuid NOT NULL REFERENCES market_users(id) ON DELETE CASCADE,
 seller_id uuid NOT NULL REFERENCES market_users(id) ON DELETE CASCADE,
 product_snapshot jsonb NOT NULL, quantity int NOT NULL CHECK(quantity>0), delivery text NOT NULL,
 address text NOT NULL DEFAULT '', status text NOT NULL DEFAULT 'Pending', total int NOT NULL CHECK(total>0),
 request_id text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(buyer_id,request_id)
);
CREATE INDEX IF NOT EXISTS market_orders_seller ON market_orders(seller_id,created_at DESC);
CREATE INDEX IF NOT EXISTS market_orders_buyer ON market_orders(buyer_id,created_at DESC);
CREATE INDEX IF NOT EXISTS market_sessions_expiry ON market_sessions(expires_at);
CREATE TABLE IF NOT EXISTS market_oidc_challenges (
 id text PRIMARY KEY, provider text NOT NULL, nonce text NOT NULL, role text NOT NULL,
 proof_hash text NOT NULL DEFAULT '', identity_email text NOT NULL DEFAULT '', expires_at timestamptz NOT NULL DEFAULT now()+interval '10 minutes'
);

-- Password accounts are unverified identities; never auto-link by email.
ALTER TABLE market_users ADD COLUMN IF NOT EXISTS password_hash text;
