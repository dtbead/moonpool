CREATE TABLE "jpeg" (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"low"	BLOB,
	"medium"	BLOB,
	"high"	BLOB,
	PRIMARY KEY("archive_id")
);

CREATE TABLE "webp" (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"low"	BLOB,
	"medium"	BLOB,
	"high"	BLOB,
	PRIMARY KEY("archive_id")
);