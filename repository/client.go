package repository

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const TimeFormat = "2006-01-02 15:04:05.000000000-07:00"

type FileInfo struct {
	Filename     string
	Path         string
	DateCreated  time.Time
	DateModified time.Time
	DateScanned  time.Time
	Author       string
	MetaData     string
}

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./fileinfo.db")
	if err != nil {
		return nil, err
	}

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
		return nil, err
	}

	return db, nil
}

func InsertFileInfo(db *sql.DB, fileInfo FileInfo) error {
	insertSQL := `INSERT INTO files (filename, path, date_created, date_modified, date_scanned, author, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, fileInfo.Filename, fileInfo.Path, fileInfo.DateCreated.Format(TimeFormat), fileInfo.DateModified.Format(TimeFormat), fileInfo.DateScanned.Format(TimeFormat), fileInfo.Author, fileInfo.MetaData)
	return err
}

func QueryFiles(db *sql.DB, filter, groupBy string) (*sql.Rows, error) {
	var query string
	if groupBy != "" {
		query = `SELECT filename, path, date_created, date_modified, date_scanned, author, metadata FROM files WHERE filename LIKE ? GROUP BY ` + groupBy
	} else {
		query = `SELECT filename, path, date_created, date_modified, date_scanned, author, metadata FROM files WHERE filename LIKE ?`
	}
	return db.Query(query, "%"+filter+"%")
}

func ScanRowToFileInfo(rows *sql.Rows) (FileInfo, error) {
	var fileInfo FileInfo
	var dateCreated, dateModified, dateScanned string
	err := rows.Scan(&fileInfo.Filename, &fileInfo.Path, &dateCreated, &dateModified, &dateScanned, &fileInfo.Author, &fileInfo.MetaData)
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
