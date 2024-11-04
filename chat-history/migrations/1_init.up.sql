CREATE TABLE IF NOT EXISTS chats
(
    id INTEGER PRIMARY KEY,
    chat_id TEXT NOT NULL,
    user_id TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user ON chats (user_id);

CREATE TABLE IF NOT EXISTS messages
(
    id INTEGER PRIMARY KEY,
    chatID TEXT NOT NULL,
    senderID TEXT NOT NULL,
    messageID TEXT NOT NULL UNIQUE,
    messageText TEXT NOT NULL,
    sentTime INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chat_id ON chats (chat_id);