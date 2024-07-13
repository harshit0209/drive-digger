package core

import (
	"database/sql"
	"drive-digger/repository"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func ScanDirectory(db *sql.DB, rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileInfo := repository.FileInfo{
				Filename:     info.Name(),
				Path:         path,
				DateCreated:  info.ModTime(), // Assuming creation date as modification time
				DateModified: info.ModTime(),
				DateScanned:  time.Now(),
				Author:       "Unknown", // Author metadata not available via os package
				MetaData:     "N/A",     // Placeholder for any additional metadata
			}
			if err := repository.InsertFileInfo(db, fileInfo); err != nil {
				return err
			}
		}
		return nil
	})
}
func QueryAndPrintFiles(db *sql.DB, filter, groupBy string) error {
	rows, err := repository.QueryFiles(db, filter, groupBy)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		fileInfo, err := repository.ScanRowToFileInfo(rows)
		if err != nil {
			return err
		}
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", fileInfo.Filename, fileInfo.Path, fileInfo.DateCreated, fileInfo.DateModified, fileInfo.DateScanned, fileInfo.Author, fileInfo.MetaData)
	}

	return nil
}

func ExportToCSV(db *sql.DB, filter, groupBy, outputFileName string) error {
	rows, err := repository.QueryFiles(db, filter, groupBy)
	if err != nil {
		return err
	}
	defer rows.Close()

	file, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Filename", "Path", "DateCreated", "DateModified", "DateScanned", "Author", "MetaData"})

	for rows.Next() {
		fileInfo, err := repository.ScanRowToFileInfo(rows)
		if err != nil {
			return err
		}
		writer.Write([]string{fileInfo.Filename, fileInfo.Path, fileInfo.DateCreated.Format(repository.TimeFormat), fileInfo.DateModified.Format(repository.TimeFormat), fileInfo.DateScanned.Format(repository.TimeFormat), fileInfo.Author, fileInfo.MetaData})
	}

	return nil
}
