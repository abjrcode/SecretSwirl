CREATE TABLE "favorite_instances" (
	"provider_code"	TEXT NOT NULL,
	"instance_id"	TEXT NOT NULL,
	PRIMARY KEY("provider_code","instance_id")
) WITHOUT ROWID;