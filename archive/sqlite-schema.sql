CREATE TABLE archive (
	"id"		integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	"path"		text NOT NULL UNIQUE,
	"extension"	TEXT
);

CREATE TABLE timestamps (
	"archive_id"	integer NOT NULL UNIQUE PRIMARY KEY,
	"date_modified"	text NOT NULL,
	"date_imported"	text NOT NULL,
	"date_created"	text NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE hashes (
	"archive_id"	integer NOT NULL UNIQUE,
	"md5"		blob NOT NULL UNIQUE,
	"sha1"		blob NOT NULL UNIQUE,
	"sha256"	blob NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);

CREATE TABLE perceptual_hashes (
	"archive_id"	integer NOT NULL,
	"hashtype"		text NOT NULL,
	"hash"		integer NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	CONSTRAINT hash_unique UNIQUE (archive_id, hashtype, hash)
);

CREATE TABLE tags (
	"tag_id"	integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	"text"		text NOT NULL UNIQUE
);

CREATE TABLE tagmap (
	"archive_id"	integer NOT NULL,
	"tag_id"	integer NOT NULL,
	FOREIGN KEY("tag_id") REFERENCES "tags"("tag_id") ON DELETE CASCADE, 
	FOREIGN KEY("archive_id") REFERENCES "archive"("id"),
	UNIQUE (archive_id, tag_id) ON CONFLICT IGNORE
);

CREATE TABLE notes (
	"archive_id"	integer NOT NULL,
	"title"		text NOT NULL,
	"text"		text NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	CONSTRAINT title_unique UNIQUE (archive_id, title)
);
