CREATE TABLE IF NOT EXISTS "aws_iam_idc" (
	"start_url"	INTEGER NOT NULL UNIQUE COLLATE NOCASE,
	"region"	TEXT NOT NULL,
	"enabled"	INTEGER NOT NULL,
	"client_id"	TEXT NOT NULL,
	"client_secret"	TEXT NOT NULL,
	"created_at"	INTEGER NOT NULL,
	"expires_at"	INTEGER NOT NULL,
	"id_token"	TEXT,
	"access_token"	TEXT,
	"token_type"	TEXT,
	"access_token_created_at"	INTEGER,
	"access_token_expires_in"	INTEGER,
	"refresh_token"	TEXT,
	PRIMARY KEY("start_url")
) WITHOUT ROWID;