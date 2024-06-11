CREATE TABLE IF NOT EXISTS "users" (
    "id"        INTEGER,
    "name"      TEXT NOT NULL CHECK(LENGTH("name") <= 256),
    "team"      TEXT NOT NULL CHECK(LENGTH("team") <= 256),
    "phone"     TEXT NOT NULL CHECK(LENGTH("phone") <= 16),
    "email"     TEXT NOT NULL CHECK(LENGTH("email") <= 256),
    "username"  TEXT,
    "time"      TEXT,
    "team_members" TEXT,
    PRIMARY KEY("id")
);


CREATE TABLE IF NOT EXISTS "blocked_users" (
	"id"	INTEGER PRIMARY KEY
);