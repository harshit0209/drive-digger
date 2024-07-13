package core

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"drive-digger/repository"

	"github.com/cheggaaa/pb/v3"
)

func ScanDirectory(db *sql.DB, rootDir string) error {
	var wg sync.WaitGroup
	fileChan := make(chan repository.FileInfo)
	doneChan := make(chan struct{})

	go func() {
		bar := pb.StartNew(0)
		for fileInfo := range fileChan {
			if err := repository.InsertFileInfo(db, fileInfo); err != nil {
				fmt.Println("Error inserting file info:", err)
			}
			bar.Increment()
		}
		bar.Finish()
		close(doneChan)
	}()

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			wg.Add(1)
			go func(info os.FileInfo, path string) {
				defer wg.Done()
				fileInfo := repository.FileInfo{
					Filename:     info.Name(),
					Path:         path,
					Size:         info.Size(),
					DateCreated:  info.ModTime(), // Assuming creation date as modification time
					DateModified: info.ModTime(),
					DateScanned:  time.Now(),
					Author:       "Unknown", // Author metadata not available via os package
					FileType:     filepath.Ext(info.Name()),
					MetaData:     "N/A", // Placeholder for any additional metadata
				}
				fileChan <- fileInfo
			}(info, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()
	close(fileChan)
	<-doneChan

	return nil
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
		fmt.Printf("%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n", fileInfo.Filename, fileInfo.Path, fileInfo.Size, fileInfo.DateCreated.Format(repository.TimeFormat), fileInfo.DateModified.Format(repository.TimeFormat), fileInfo.DateScanned.Format(repository.TimeFormat), fileInfo.Author, fileInfo.FileType, fileInfo.MetaData)
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
	writer.Write([]string{"Filename", "Path", "Size", "DateCreated", "DateModified", "DateScanned", "Author", "FileType", "MetaData"})

	for rows.Next() {
		fileInfo, err := repository.ScanRowToFileInfo(rows)
		if err != nil {
			return err
		}
		writer.Write([]string{fileInfo.Filename, fileInfo.Path, fmt.Sprintf("%d", fileInfo.Size), fileInfo.DateCreated.Format(repository.TimeFormat), fileInfo.DateModified.Format(repository.TimeFormat), fileInfo.DateScanned.Format(repository.TimeFormat), fileInfo.Author, fileInfo.FileType, fileInfo.MetaData})
	}

	return nil
}
