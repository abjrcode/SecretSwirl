CREATE TABLE IF NOT EXISTS "aws_credentials_file" (
	"instance_id"	TEXT NOT NULL UNIQUE COLLATE NOCASE,
	"version" INTEGER NOT NULL,
	"file_path"	TEXT NOT NULL,
	"aws_profile_name" TEXT NOT NULL,
	"label"	TEXT NOT NULL,
	"provider_code"	TEXT NOT NULL,
	"provider_id"	TEXT NOT NULL,
	"created_at"	INTEGER NOT NULL,
	"last_drained_at"	INTEGER,
	PRIMARY KEY("instance_id")
) WITHOUT ROWID;