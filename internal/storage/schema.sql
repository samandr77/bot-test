-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY, -- Telegram UserID
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица балансов
CREATE TABLE IF NOT EXISTS balances (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    gpt_credits INT DEFAULT 0,
    sora_credits INT DEFAULT 0,
    nanobanana_credits INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица истории генераций
CREATE TABLE IF NOT EXISTS generation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    model_type TEXT NOT NULL, -- 'gpt', 'sora', 'nanobanana'
    prompt TEXT,
    media_url TEXT, -- Ссылка на сгенерированный файл
    created_at TIMESTAMPTZ DEFAULT NOW()
);
