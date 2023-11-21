CREATE TABLE "providers" (
	"code"	TEXT NOT NULL,
	"instance_id"	TEXT NOT NULL,
	"display_name"	TEXT NOT NULL,
	"is_favorite"	INTEGER NOT NULL,
	PRIMARY KEY("code","instance_id")
) WITHOUT ROWID;

CREATE TABLE "argon_key_material" (
	"key_id"	TEXT NOT NULL UNIQUE,
	"key_hash_sha3_512"	TEXT NOT NULL,
  "argon2_version"  TEXT NOT NULL,
  "argon2_variant"  TEXT NOT NULL,
  "created_at"  INTEGER NOT NULL,
	"memory"	INTEGER NOT NULL,
	"iterations"	INTEGER NOT NULL,
	"parallelism"	INTEGER NOT NULL,
	"salt_length"	INTEGER NOT NULL,
	"salt_base64"	TEXT NOT NULL,
	"key_length"	INTEGER NOT NULL,
	PRIMARY KEY("key_id")
) WITHOUT ROWID;