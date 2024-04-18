package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/file"
	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/media"
)

const (
	defaultDatabaseName = "archive"
	defaultDatabaseType = "sqlite3"
)

var (
	ImportCMD      = flag.NewFlagSet("import", flag.ExitOnError)
	ImportFilename = ImportCMD.String("file", "", "direct path of a file to import")
	ImportTags     = ImportCMD.String("tags", "", "tags to import alongside file. comma separated")
	ImportDatabase = ImportCMD.String("database", fmt.Sprintf("%s.%s", defaultDatabaseName, defaultDatabaseType), "database file path to interface with")

	SearchCMD      = flag.NewFlagSet("search", flag.ExitOnError)
	SearchTag      = SearchCMD.String("tag", "", "tag to search for")
	SearchDatabase = SearchCMD.String("database", fmt.Sprintf("%s.%s", defaultDatabaseName, defaultDatabaseType), "database file to search from")

	CreateCMD          = flag.NewFlagSet("create", flag.ExitOnError)
	CreateDatabaseType = CreateCMD.String("type", defaultDatabaseType, "database type to create")
	CreateDatabaseName = CreateCMD.String("name", fmt.Sprintf("%s.%s", defaultDatabaseName, defaultDatabaseType), "path to create database")

	defaultDatabaseMedia string
	l                    = log.NewSlogLogger(context.Background())
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic("unable to get current working directory")
	}
	defaultDatabaseMedia = fmt.Sprintf("%s/media/", strings.ReplaceAll(wd, "\\", "/")) // wangblows 10 microshaft

}

func main() {
	flag.CommandLine.SetOutput(os.Stdout)

	if os.Args[1] == "-h" || os.Args[1] == "-help" {
		fmt.Println("use 'import | search | create -h' for commands")
		os.Exit(1)
	}

	if len(os.Args) <= 1 {
		flag.Usage()
		fmt.Println("no arguments given. see 'import | search | create -h' for commands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "import":
		if err := ImportCMD.Parse(os.Args[2:]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		archive, err := db.OpenSQLite3(*ImportDatabase, l)
		if err != nil {
			slog.Error("unable to open database. %v", err)
			os.Exit(1)
		}

		f, err := os.Open(*ImportFilename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		archive.TXBegin()
		if err := Import(*f, strings.Split(*ImportTags, ","), archive); err != nil {
			fmt.Println(err)
			archive.TXRollback()
			os.Exit(1)
		}

		archive.TXCommit()
	case "create":
		CreateCMD.Parse(os.Args[2:])
		if *CreateDatabaseType != "sqlite3" {
			fmt.Println("only sqlite3 is supported as a database")
			os.Exit(1)
		}
		if err := CreateArchive(*CreateDatabaseName); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "search":
		SearchCMD.Parse(os.Args[2:])
		s, err := db.OpenSQLite3(*SearchDatabase, l)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		res, err := s.SearchTag(*SearchTag)
		if err != nil {
			slog.Error(err.Error())
		}

		for i := 0; i < len(res); i++ {
			fmt.Println(res[i].Metadata.PathRelative)
		}
	default:
		fmt.Printf("unknown command '%s'\n", os.Args[1])
	}
}

func CreateArchive(filepath string) error {
	archive, err := db.NewSQLite3(filepath, l)
	archive.Initialize()
	if err != nil {
		return err
	}
	return archive.Close()
}

// Import imports a new entry into an archive
func Import(f os.File, tags []string, archive db.Database) error {
	e, err := file.CopyAndHash(defaultDatabaseMedia, path.Ext(f.Name()), &f)
	if err != nil {
		return err
	}

	dm, err := file.GetDateModified(&f)
	if err != nil {
		return err
	}

	e.Metadata.Timestamp.DateModifiedUTC = dm.UTC()
	e.Metadata.Timestamp.DateImportedUTC = time.Now().UTC()

	archiveID, err := archive.InsertEntry(media.Hashes(e.Metadata.Hash), e.Metadata.PathRelative, e.Metadata.Extension)
	if err != nil {
		return err
	}

	_, err = archive.AddTags(tags)
	if err != nil {
		return err
	}

	archive.MapTags(archiveID, tags)

	archive.SetTimestamp(archiveID, e.Metadata.Timestamp)

	return nil
}
