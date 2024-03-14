package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/file"
)

const (
	testdb              = "E:/Programming/go/src/github.com/dtbead/moonpool/db/testdata/prefilled.sqlite3"
	defaultGroupName    = "archive"
	defaultDatabaseType = "sqlite3"
	defaultDatabaseName = defaultGroupName
)

var (
	ImportCMD       = flag.NewFlagSet("import", flag.ExitOnError)
	ImportFileName  = ImportCMD.String("file", "", "direct path of a file to import")
	ImportGroupName = ImportCMD.String("group", defaultGroupName, "group to import into")
	ImportDatabase  = ImportCMD.String("database", "db.sqlite3", "database file path to interface with")
	// importRootDirectory := importCMD.String("root", "", "root directory to media storage")

	SearchCMD       = flag.NewFlagSet("search", flag.ExitOnError)
	SearchTag       = SearchCMD.String("tag", "", "tag to search for")
	SearchGroupName = SearchCMD.String("group", defaultGroupName, "group to search from")
	SearchDatabase  = SearchCMD.String("database", "db.sqlite3", "database file to search from")

	CreateCMD      = flag.NewFlagSet("create", flag.ExitOnError)
	CreateDBType   = CreateCMD.String("type", defaultDatabaseType, "database type to create")
	CreateFileName = CreateCMD.String("name", fmt.Sprintf("%s.%s", defaultGroupName, defaultDatabaseType), "path to create database")

	defaultDatabaseMedia string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic("unable to get current working directory")
	}
	defaultDatabaseMedia = fmt.Sprintf("%s/media", wd)
}

func main() {
	flag.Parse()

	ImportCMD.SetOutput(os.Stdout)
	CreateCMD.SetOutput(os.Stdout)
	SearchCMD.SetOutput(os.Stdout)

	if len(os.Args) <= 1 {
		fmt.Println("no arguments given")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "import":
		ImportCMD.Parse(os.Args[2:])

		db, err := db.OpenSQLite3(*ImportDatabase)
		if err != nil {
			slog.Error("unable to open database. %v", err)
			os.Exit(1)
		}

		f, err := os.Open(*ImportFileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		e := BuildEntry(f)
		if err := Import(*ImportGroupName, e, db); err != nil { // fails to copy file to storage
			fmt.Println(err)
			os.Exit(1)
		}
	case "create":
		CreateCMD.Parse(os.Args[2:])
		if *CreateDBType != "sqlite3" {
			fmt.Println("only sqlite3 is supported as a database")
			os.Exit(1)
		}
		if err := CreateArchive(*CreateFileName); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "search":
		SearchCMD.Parse(os.Args[2:])
		s, err := db.OpenSQLite3(*SearchDatabase)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		res, err := s.SearchTag("archive", *SearchTag) // hardcoded defaults should be removed
		if err != nil {
			slog.Error(err.Error())
		}

		for i := 0; i < len(res); i++ {
			fmt.Println(res[i].Metadata.Path)
		}
	default:
		fmt.Printf("unknown command '%s'\n", os.Args[1])
	}
}

func CreateArchive(filepath string) error {
	d, err := db.NewDatabase(filepath)
	if err != nil {
		return err
	}
	return d.Close()
}

func BuildEntry(f *os.File) file.Entry {
	var e file.Entry
	e.File = *f
	e.Metadata.MD5Hash = file.GetMD5Hash(f)

	e.Metadata.Path = file.BuildFilePath(defaultDatabaseMedia, file.ByteToString(e.Metadata.MD5Hash), path.Ext(e.File.Name()))

	ts, err := file.GetTimestamp(f)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get timestamp for %s. %v", f.Name(), err))
	}
	e.Metadata.Timestamp = ts

	return e
}

// Import imports a new entry into an archive
func Import(table string, e file.Entry, d db.Database) error {
	// tags
	if err := d.AddTags(e.Tags); err != nil {
		slog.Error(fmt.Sprintf("failed to insert tags to %s. %v", table, err))
	}

	// hash, path, ext
	if err := d.InsertToArchive(table, string(e.Metadata.MD5Hash), e.Metadata.Path, path.Ext(e.Metadata.Path)); err != nil {
		//this entire error handling sucks but im not sure how to do it any better yet
		return err
	}

	// file
	if err := file.Copy(e.Metadata.Path, &e.File); err != nil {
		errmsg := fmt.Sprintf("failed to copy '%s' to storage. %v", e.File.Name(), err)
		slog.Error(errmsg)
		return errors.New(errmsg)
	}

	return nil
}
