package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/elves/elvish/store"
	_ "github.com/mattn/go-sqlite3" // enable the "sqlite3" SQL driver
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <db_migrate> -f <old-db> -t <new-db>")
		return
	}

	sdb, err := openSqlite(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	s, err := store.NewStore(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	log.Print("start migrating...")
	if err := migrateCmd(sdb, s); err != nil {
		log.Fatal(err)
	}
	if err := migrateDir(sdb, s); err != nil {
		log.Fatal(err)
	}
	if err := migrateVar(sdb, s); err != nil {
		log.Fatal(err)
	}

	sdb.Close()
	s.Close()

	log.Print("done")
}

func migrateCmd(sdb *sql.DB, s *store.Store) error {
	rows, err := sdb.Query(`SELECT content FROM cmd ORDER BY rowid`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		cmd string
		cnt uint64
	)

	for rows.Next() {
		if err := rows.Scan(&cmd); err != nil {
			return err
		}
		if _, err := s.AddCmd(cmd); err != nil {
			return err
		}
		cnt += 1
	}
	log.Printf("%d commands migrated", cnt)
	return nil
}

func migrateDir(sdb *sql.DB, s *store.Store) error {
	rows, err := sdb.Query(`SELECT path, score FROM dir`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		path  string
		score float64
		cnt   uint64
	)

	for rows.Next() {
		if err := rows.Scan(&path, &score); err != nil {
			return err
		}
		if err := s.AddDirRaw(path, score); err != nil {
			return err
		}
		cnt += 1
	}
	log.Printf("%d directories migrated", cnt)
	return nil
}

func migrateVar(sdb *sql.DB, s *store.Store) error {
	rows, err := sdb.Query(`SELECT name, value FROM shared_var`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		name  string
		value string
		cnt   uint64
	)

	for rows.Next() {
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}
		if err := s.SetSharedVar(name, value); err != nil {
			return err
		}
		cnt += 1
	}
	log.Printf("%d shared variables migrated", cnt)
	return nil
}

func openSqlite(path string) (*sql.DB, error) {
	uri := "file:" + url.QueryEscape(path) +
		"?mode=rwc&cache=shared&vfs=unix-dotfile"
	db, err := sql.Open("sqlite3", uri)
	if err == nil {
		db.SetMaxOpenConns(1)
	}

	return db, err
}
