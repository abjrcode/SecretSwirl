CREATE TABLE IF NOT EXISTS "aws_sso_clients" (
	"client_id"	TEXT NOT NULL,
	"client_secret_enc"	BLOB NOT NULL,
	"created_at"	INTEGER NOT NULL,
	"expires_at"	INTEGER NOT NULL,
	"enc_key_id"	TEXT NOT NULL,
	PRIMARY KEY("client_id")
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS "aws_idc" (
	"instance_id"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"version" INTEGER NOT NULL,
	"label"	TEXT NOT NULL,
	"start_url"	INTEGER NOT NULL UNIQUE COLLATE NOCASE,
	"region"	TEXT NOT NULL,
	"enabled"	INTEGER NOT NULL,
	"id_token_enc"	TEXT NOT NULL,
	"access_token_enc"	BLOB NOT NULL,
	"token_type"	TEXT NOT NULL,
	"access_token_created_at"	INTEGER NOT NULL,
	"access_token_expires_in"	INTEGER NOT NULL,
	"refresh_token_enc"	BLOB NOT NULL,
	"enc_key_id"	TEXT NOT NULL,
	PRIMARY KEY("instance_id")
) WITHOUT ROWID;