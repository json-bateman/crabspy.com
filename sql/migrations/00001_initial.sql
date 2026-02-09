-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE rooms (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    host_id INTEGER NOT NULL REFERENCES users(id),
    max_locations INTEGER NOT NULL DEFAULT 30,
    max_players INTEGER NOT NULL DEFAULT 8,
    status TEXT NOT NULL DEFAULT 'lobby' CHECK(status IN ('lobby', 'in_progress', 'finished')),
    is_private INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE room_members (
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at INTEGER NOT NULL DEFAULT (unixepoch()),
    is_ready INTEGER NOT NULL DEFAULT 0,
    team INTEGER,
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_room_members_user ON room_members(user_id);
