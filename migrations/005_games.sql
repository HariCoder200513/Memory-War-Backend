CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(30) NOT NULL,
    description VARCHAR(120) DEFAULT '',
    category VARCHAR(50) NOT NULL,
    grid_rows INTEGER NOT NULL DEFAULT 4,
    grid_columns INTEGER NOT NULL DEFAULT 4,
    card_icons TEXT[] DEFAULT '{}',
    card_back VARCHAR(20) NOT NULL DEFAULT 'classic',
    time_limit_seconds INTEGER NOT NULL DEFAULT 120,
    move_limit INTEGER NOT NULL DEFAULT 0,
    mismatch_delay_ms INTEGER NOT NULL DEFAULT 1000,
    allow_hints BOOLEAN NOT NULL DEFAULT true,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_games_creator_id ON games(creator_id);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
