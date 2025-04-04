CREATE TABLE archive (
	"id"		INTEGER PRIMARY KEY AUTOINCREMENT,
	"path"		TEXT NOT NULL UNIQUE,
	"extension"	TEXT
);

CREATE TABLE archive_timestamps (
	"archive_id"	INTEGER PRIMARY KEY,
	"date_modified"	INTEGER NOT NULL,
	"date_imported"	INTEGER NOT NULL,
	"date_created"	INTEGER NOT NULL,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);

CREATE TABLE "archive_metadata" (
	"archive_id"	INTEGER PRIMARY KEY,
	"file_size"	INTEGER NOT NULL,
	"file_mimetype"	TEXT NOT NULL DEFAULT "unknown",
	"media_width"  INTEGER,
	"media_height" INTEGER,
	"media_orientation" TEXT,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);

CREATE TABLE hashes_chksum (
	"archive_id"	INTEGER PRIMARY KEY,
	"md5"			BLOB NOT NULL UNIQUE CHECK (length(md5) == 16),
	"sha1"			BLOB NOT NULL UNIQUE CHECK (length(sha1) == 20),
	"sha256"		BLOB NOT NULL UNIQUE CHECK (length(sha256) == 32),
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);

CREATE TABLE hashes_perceptual (
	"archive_id"	INTEGER NOT NULL,
	"hash_type"		TEXT NOT NULL,
	"hash"			INTEGER NOT NULL,
	PRIMARY KEY (archive_id, hash_type),
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE
);

CREATE TABLE tags (
	"tag_id"	INTEGER PRIMARY KEY AUTOINCREMENT,
	"text"		TEXT NOT NULL UNIQUE
);

CREATE TABLE tags_alias (
	"tag_id"	INTEGER,
	"text"		TEXT NOT NULL,
	PRIMARY KEY (tag_id, text)
	FOREIGN KEY("tag_id") REFERENCES "tags"("tag_id") ON DELETE CASCADE
);

CREATE TABLE tag_map (
	"tag_id"	INTEGER NOT NULL,
	"archive_id"		INTEGER NOT NULL,
	FOREIGN KEY("tag_id") REFERENCES "tags"("tag_id") ON DELETE CASCADE, 
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	UNIQUE (tag_id, archive_id) ON CONFLICT IGNORE
);

CREATE TABLE tag_count (
	"tag_id"	INTEGER NOT NULL UNIQUE PRIMARY KEY,
	"total"		INTEGER NOT NULL DEFAULT 1,
	FOREIGN KEY("tag_id") REFERENCES "tags"("tag_id")  ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE notes (
	"archive_id"	INTEGER NOT NULL UNIQUE PRIMARY KEY,
	"title"			TEXT NOT NULL,
	"text"			TEXT NOT NULL UNIQUE,
	FOREIGN KEY("archive_id") REFERENCES "archive"("id") ON DELETE CASCADE,
	CONSTRAINT unique_title UNIQUE (archive_id, title)
) WITHOUT ROWID;

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
