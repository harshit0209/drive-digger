package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type FileInfo struct {
	Filename     string
	Path         string
	DateCreated  time.Time
	DateModified time.Time
	DateScanned  time.Time
	Author       string
	MetaData     string
}

func main() {
	db, err := sql.Open("sqlite3", "./fileinfo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        filename TEXT,
        path TEXT,
        date_created TEXT,
        date_modified TEXT,
        date_scanned TEXT,
        author TEXT,
        metadata TEXT
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	rootDir := "./" // Change this to the root directory you want to scan
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileInfo := FileInfo{
				Filename:     info.Name(),
				Path:         path,
				DateCreated:  info.ModTime(), // Assuming creation date as modification time
				DateModified: info.ModTime(),
				DateScanned:  time.Now(),
				Author:       "Unknown", // Author metadata not available via os package
				MetaData:     "N/A",     // Placeholder for any additional metadata
			}
			insertFileInfo(db, fileInfo)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File information inserted into database.")
}

func insertFileInfo(db *sql.DB, fileInfo FileInfo) {
	insertSQL := `INSERT INTO files (filename, path, date_created, date_modified, date_scanned, author, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, fileInfo.Filename, fileInfo.Path, fileInfo.DateCreated, fileInfo.DateModified, fileInfo.DateScanned, fileInfo.Author, fileInfo.MetaData)
	if err != nil {
		log.Fatal(err)
	}
}
