CREATE TABLE archive (
	"id"		INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	"path"		TEXT NOT NULL UNIQUE,
	"extension"	TEXT
);

CREATE TABLE archive_timestamps (
	"archive_id"	INTEGER NOT NULL UNIQUE PRIMARY KEY,
	"date_modified"	TEXT NOT NULL,
	"date_imported"	TEXT NOT NULL,
	"date_created"	text NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE hashes_chksum (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"md5"			BLOB NOT NULL UNIQUE,
	"sha1"			BLOB NOT NULL UNIQUE,
	"sha256"		BLOB NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);

CREATE TABLE hashes_perceptual (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"hash_type"		TEXT NOT NULL,
	"hash"			INTEGER NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	CONSTRAINT hash_unique UNIQUE (archive_id, hash_type, hash)
);

CREATE TABLE tags (
	"tag_id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	"text"		TEXT NOT NULL UNIQUE
);

CREATE TABLE tag_map (
	"archive_id"	INTEGER NOT NULL UNIQUE,
	"tag_id"		INTEGER NOT NULL,
	FOREIGN KEY("tag_id") REFERENCES "tags"("tag_id") ON DELETE CASCADE, 
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	UNIQUE (archive_id, tag_id) ON CONFLICT IGNORE
);

CREATE TABLE tag_count (
	"tag_id"	INTEGER NOT NULL UNIQUE,
	"total"		INTEGER NOT NULL DEFAULT 1,
	FOREIGN KEY("tag_id") REFERENCES "tags"("tag_id")
);

CREATE TABLE notes (
	"archive_id"	INTEGER NOT NULL,
	"title"			TEXT NOT NULL,
	"text"			TEXT NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	CONSTRAINT unique_title UNIQUE (archive_id, title)
);

CREATE TRIGGER tags_update_count AFTER INSERT ON tag_map 
BEGIN	
	INSERT INTO tag_count(tag_id, total) VALUES(NEW.tag_id, 1)
	ON CONFLICT(tag_id)
	DO
		UPDATE SET total = total + 1; 
END;

CREATE TRIGGER tags_remove_count AFTER DELETE ON tag_map 
BEGIN
		UPDATE tag_count
		SET total = total - 1
		WHERE tag_id == OLD.tag_id;
END;
