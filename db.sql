CREATE TABLE "users" (
	"id"	INTEGER,
	"name"	TEXT NOT NULL CHECK(LENGTH("name") <= 256),
	"team"	TEXT NOT NULL CHECK(LENGTH("team") <= 256),
	"phone"	TEXT NOT NULL CHECK(LENGTH("phone") <= 16),
	"email"	TEXT NOT NULL CHECK(LENGTH("email") <= 256),
	"username"	TEXT,
	"time"	TEXT,
	PRIMARY KEY("id")
);