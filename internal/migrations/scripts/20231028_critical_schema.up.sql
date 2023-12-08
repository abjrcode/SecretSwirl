CREATE TABLE "argon_keys" (
	"key_id"	TEXT NOT NULL UNIQUE,
	"version"	INTEGER NOT NULL,
	"key_hash_sha3_512"	BLOB NOT NULL,
	"argon2_version"	TEXT NOT NULL,
	"argon2_variant"	TEXT NOT NULL,
	"created_at"	INTEGER NOT NULL,
	"memory"	INTEGER NOT NULL,
	"iterations"	INTEGER NOT NULL,
	"parallelism"	INTEGER NOT NULL,
	"salt_length"	INTEGER NOT NULL,
	"salt_base64"	TEXT NOT NULL,
	"key_length"	INTEGER NOT NULL,
	PRIMARY KEY("key_id")
) WITHOUT ROWID;

CREATE TABLE "event_log" (
	"id"	INTEGER NOT NULL UNIQUE,
	"event_type"	TEXT NOT NULL,
	"event_version" INTEGER NOT NULL,
	"data"	BLOB NOT NULL,
	"source_type"	TEXT NOT NULL,
	"source_id"	TEXT NOT NULL,
	"user_id"	TEXT NOT NULL,
	"created_at"	INTEGER NOT NULL,
	"causation_id"	TEXT NOT NULL,
	"correlation_id"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	UNIQUE("source_id", "event_version")
);