-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    crab_avatar TEXT NOT NULL DEFAULT 'tourist-crab.png'
);

CREATE TABLE rooms (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    code TEXT UNIQUE NOT NULL UNIQUE,
    host_id INTEGER NOT NULL REFERENCES users(id),
    max_locations INTEGER NOT NULL DEFAULT 30,
    max_players INTEGER NOT NULL DEFAULT 8,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    state TEXT NOT NULL DEFAULT 'lobby' CHECK (state IN ('lobby', 'game'))
);

CREATE TABLE room_members (
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at INTEGER NOT NULL DEFAULT (unixepoch()),
    is_ready INTEGER NOT NULL DEFAULT 0,
    points INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE games (
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    spy_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    paused INTEGER NOT NULL DEFAULT 0,
    started_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (room_id)
);

CREATE INDEX idx_room_members_user ON room_members(user_id);
