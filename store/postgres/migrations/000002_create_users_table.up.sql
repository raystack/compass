CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    uuid text,
    email text,
    provider text,
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW()
);

CREATE UNIQUE INDEX users_idx_uuid ON users(uuid);
CREATE UNIQUE INDEX users_idx_email ON users(email);