package cmd

import (
	"bufio"
	"drive-digger/core"
	"drive-digger/repository"
	"fmt"
	"os"
	"strings"

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
		color.Cyan("What do you want to do? (scan, query, export, exit)")
		fmt.Print("Enter choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "scan":
			fmt.Print("Enter directory to scan: ")
			rootDir, _ := reader.ReadString('\n')
			rootDir = strings.TrimSpace(rootDir)
			color.Cyan("Scanning directory: %s", rootDir)
			err := core.ScanDirectory(db, rootDir)
			if err != nil {
				color.Red("Failed to scan directory: %v", err)
				continue
			}
			color.Green("File information inserted into database.")
		case "query":
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
				continue
			}
		case "export":
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
				continue
			}
			color.Green("Data exported to CSV file: %s", outputFileName)
		case "exit":
			color.Green("Exiting program.")
			return
		default:
			color.Yellow("Invalid choice. Please enter scan, query, export, or exit.")
		}
	}
}
