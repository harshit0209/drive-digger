package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"drive-digger/core"
	"drive-digger/repository"

	"github.com/fatih/color"
)

func Run() {
	db, err := repository.OpenDB()
	if err != nil {
		color.Red("Failed to open database: %v", err)
		return
	}
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		color.Cyan("Main Menu:\n1. Files Analysis\n2. Files Operation\n3. Exit")
		fmt.Print("Enter choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			handleFilesAnalysis(db, reader)
		case "2":
			handleFilesOperation(db, reader)
		case "3":
			color.Green("Exiting program.")
			return
		default:
			color.Yellow("Invalid choice. Please enter 1, 2, or 3.")
		}
	}
}

func handleFilesAnalysis(db *sql.DB, reader *bufio.Reader) {
	for {
		color.Cyan("Files Analysis:\n1. Scan Directory\n2. Query Files\n3. Export Files to CSV\n4. Back to Main Menu")
		fmt.Print("Enter choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter directory to scan: ")
			rootDir, _ := reader.ReadString('\n')
			rootDir = strings.TrimSpace(rootDir)
			color.Cyan("Scanning directory: %s", rootDir)
			err := core.ScanDirectory(db, rootDir)
			if err != nil {
				color.Red("Failed to scan directory: %v", err)
			} else {
				color.Green("File information inserted into database.")
			}
		case "2":
			fmt.Print("Enter filter: ")
			filter, _ := reader.ReadString('\n')
			filter = strings.TrimSpace(filter)
			fmt.Print("Enter group by field (optional): ")
			groupBy, _ := reader.ReadString('\n')
			groupBy = strings.TrimSpace(groupBy)
			color.Cyan("Querying files with filter: %s and group by: %s", filter, groupBy)
			err := core.QueryAndPrintFiles(db, filter, groupBy)
			if err != nil {
				color.Red("Failed to query files: %v", err)
			}
		case "3":
			fmt.Print("Enter filter: ")
			filter, _ := reader.ReadString('\n')
			filter = strings.TrimSpace(filter)
			fmt.Print("Enter group by field (optional): ")
			groupBy, _ := reader.ReadString('\n')
			groupBy = strings.TrimSpace(groupBy)
			fmt.Print("Enter output CSV file name: ")
			outputFileName, _ := reader.ReadString('\n')
			outputFileName = strings.TrimSpace(outputFileName)
			color.Cyan("Exporting files to CSV with filter: %s and group by: %s", filter, groupBy)
			err := core.ExportToCSV(db, filter, groupBy, outputFileName)
			if err != nil {
				color.Red("Failed to export files to CSV: %v", err)
			} else {
				color.Green("Data exported to CSV file: %s", outputFileName)
			}
		case "4":
			return
		default:
			color.Yellow("Invalid choice. Please enter 1, 2, 3, or 4.")
		}
	}
}

func handleFilesOperation(db *sql.DB, reader *bufio.Reader) {
	for {
		color.Cyan("Files Operation:\n1. Structured Copy Media Files\n2. Back to Main Menu")
		fmt.Print("Enter choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter output CSV file name: ")
			outputFileName, _ := reader.ReadString('\n')
			outputFileName = strings.TrimSpace(outputFileName)
			color.Cyan("Copying media files from the database")
			err := core.StructuredCopyMediaFiles(db, outputFileName)
			if err != nil {
				color.Red("Failed to copy media files: %v", err)
			} else {
				color.Green("Media files copied and file operations exported to CSV file: %s", outputFileName)
			}
		case "2":
			return
		default:
			color.Yellow("Invalid choice. Please enter 1 or 2.")
		}
	}
}
