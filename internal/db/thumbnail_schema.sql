CREATE TABLE "thumbnail" (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"has_jpeg"	INTEGER NOT NULL,
	"has_webp"	INTEGER NOT NULL,
	PRIMARY KEY("archive_id")
);

CREATE TABLE "thumbnail_jpeg" (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"small"	BLOB,
	"medium"	BLOB,
	"large"	BLOB,
	FOREIGN KEY("archive_id") REFERENCES "thumbnail"("archive_id") ON DELETE CASCADE
);

CREATE TABLE "thumbnail_blurhash" (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"hash" TEXT NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "thumbnail"("archive_id") ON DELETE CASCADE
);