package core

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"drive-digger/repository"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
)

const numWorkers = 10

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

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(db, fileChan, &wg)
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			color.Red("Failed to scan directory: %v", err)
			return nil
		}
		if !info.IsDir() {
			fileInfo := repository.FileInfo{
				Filename:     info.Name(),
				Path:         path,
				Size:         info.Size(),
				DateCreated:  info.ModTime(), // Assuming creation date as modification time
				DateModified: info.ModTime(),
				DateScanned:  time.Now(),
				Author:       "Unknown", // Author metadata not available via os package
				FileType:     strings.ToLower(filepath.Ext(info.Name())),
				MetaData:     "N/A", // Placeholder for any additional metadata
			}
			fileChan <- fileInfo
		}
		return nil
	})

	if err != nil {
		return err
	}

	close(fileChan)
	wg.Wait()
	<-doneChan

	return nil
}

func worker(db *sql.DB, tasks <-chan repository.FileInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	for fileInfo := range tasks {
		if err := repository.InsertFileInfo(db, fileInfo); err != nil {
			fmt.Println("Error inserting file info:", err)
		}
	}
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

func StructuredCopyMediaFiles(db *sql.DB) error {
	mediaExtensions := map[string]string{
		".jpg": "Image", ".jpeg": "Image", ".png": "Image", ".gif": "Image",
		".mp4": "Video", ".avi": "Video", ".mov": "Video", ".mkv": "Video",
		".mp3": "Audio", ".wav": "Audio", ".flac": "Audio",
	}

	var wg sync.WaitGroup
	fileChan := make(chan repository.FileInfo)
	doneChan := make(chan struct{})

	go func() {
		bar := pb.StartNew(0)
		for fileInfo := range fileChan {
			year := fileInfo.DateModified.Year()
			category, exists := mediaExtensions[strings.ToLower(fileInfo.FileType)]
			if exists {
				destDir := filepath.Join(fmt.Sprintf("%d", year), category)
				destPath := filepath.Join(destDir, fileInfo.Filename)
				os.MkdirAll(destDir, os.ModePerm)
				err := copyFile(fileInfo.Path, destPath)
				if err == nil {
					op := repository.FileOperation{
						FromPath: fileInfo.Path,
						ToPath:   destPath,
						Size:     fileInfo.Size,
						Date:     time.Now(),
					}
					if err := repository.InsertFileOperation(db, op); err != nil {
						fmt.Println("Error inserting file operation:", err)
					}
				} else {
					fmt.Println("Error copying file:", err)
				}
			}
			bar.Increment()
		}
		bar.Finish()
		close(doneChan)
	}()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(db, fileChan, &wg)
	}

	rows, err := repository.QueryFiles(db, "", "")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		fileInfo, err := repository.ScanRowToFileInfo(rows)
		if err != nil {
			return err
		}
		fileChan <- fileInfo
	}

	close(fileChan)
	wg.Wait()
	<-doneChan

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
