CREATE TABLE archive (
	"id"		integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	"path"		text NOT NULL UNIQUE,
	"extension"	TEXT
);

CREATE TABLE timestamps (
	"archive_id"	integer NOT NULL UNIQUE PRIMARY KEY,
	"date_modified"	text NOT NULL,
	"date_imported"	text NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE hashes (
	"archive_id"	integer NOT NULL UNIQUE,
	"md5"		blob NOT NULL UNIQUE,
	"sha1"		blob NOT NULL UNIQUE,
	"sha256"	blob NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
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
	UNIQUE (archive_id,  tag_id) ON CONFLICT IGNORE
);

CREATE TABLE description (
	"archive_id"	integer NOT NULL,
	"text"		text NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);
