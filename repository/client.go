package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const TimeFormat = "2006-01-02 15:04:05.000000000-07:00"

type FileInfo struct {
	Filename     string
	Path         string
	Size         int64
	DateCreated  time.Time
	DateModified time.Time
	DateScanned  time.Time
	Author       string
	FileType     string
	MetaData     string
}

type FileOperation struct {
	FromPath     string
	ToPath       string
	Size         int64
	Date         time.Time
	Filename     string
	FileType     string
	DateModified time.Time
	DateScanned  time.Time
	MetaData     string
}

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./fileinfo.db")
	if err != nil {
		return nil, err
	}

	createFileInfoTableSQL := `CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT,
		path TEXT,
		size INTEGER,
		date_created TEXT,
		date_modified TEXT,
		date_scanned TEXT,
		author TEXT,
		file_type TEXT,
		metadata TEXT
	);`

	_, err = db.Exec(createFileInfoTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func InsertFileInfo(db *sql.DB, fileInfo FileInfo) error {
	metadata := fmt.Sprintf("Filename: %s, Path: %s, Size: %d, DateCreated: %s, DateModified: %s, DateScanned: %s, Author: %s, FileType: %s",
		fileInfo.Filename, fileInfo.Path, fileInfo.Size, fileInfo.DateCreated.Format(TimeFormat), fileInfo.DateModified.Format(TimeFormat), fileInfo.DateScanned.Format(TimeFormat), fileInfo.Author, fileInfo.FileType)

	insertSQL := `INSERT INTO files (filename, path, size, date_created, date_modified, date_scanned, author, file_type, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, fileInfo.Filename, fileInfo.Path, fileInfo.Size, fileInfo.DateCreated.Format(TimeFormat), fileInfo.DateModified.Format(TimeFormat), fileInfo.DateScanned.Format(TimeFormat), fileInfo.Author, fileInfo.FileType, metadata)
	return err
}

func QueryFiles(db *sql.DB, filter, groupBy string) (*sql.Rows, error) {
	var query string
	if groupBy != "" {
		query = `SELECT filename, path, size, date_created, date_modified, date_scanned, author, file_type, metadata FROM files WHERE filename LIKE ? GROUP BY ` + groupBy
	} else {
		query = `SELECT filename, path, size, date_created, date_modified, date_scanned, author, file_type, metadata FROM files WHERE filename LIKE ?`
	}
	return db.Query(query, "%"+filter+"%")
}

func ScanRowToFileInfo(rows *sql.Rows) (FileInfo, error) {
	var fileInfo FileInfo
	var dateCreated, dateModified, dateScanned string
	err := rows.Scan(&fileInfo.Filename, &fileInfo.Path, &fileInfo.Size, &dateCreated, &dateModified, &dateScanned, &fileInfo.Author, &fileInfo.FileType, &fileInfo.MetaData)
	if err != nil {
		return fileInfo, err
	}

	// Convert date strings to time.Time
	fileInfo.DateCreated, err = time.Parse(TimeFormat, dateCreated)
	if err != nil {
		return fileInfo, err
	}
	fileInfo.DateModified, err = time.Parse(TimeFormat, dateModified)
	if err != nil {
		return fileInfo, err
	}
	fileInfo.DateScanned, err = time.Parse(TimeFormat, dateScanned)
	if err != nil {
		return fileInfo, err
	}

	return fileInfo, nil
}
