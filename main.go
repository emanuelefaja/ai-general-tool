package main

import (
	"fmt"
	"os"

	"ai-general-tool/tools"
)

func printUsage() {
	fmt.Println("AI General Tool - Data Enrichment Toolkit")
	fmt.Println()
	fmt.Println("Usage: go run . <command> [flags]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("DATA INPUT:")
	fmt.Println("  read-csv      Read and analyze a CSV file")
	fmt.Println("  read-excel    Read and analyze an Excel file")
	fmt.Println()
	fmt.Println("DATA PROCESSING:")
	fmt.Println("  process-data  Process data with AI to add new columns")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . read-csv data.csv")
	fmt.Println("  go run . read-csv data.csv -rows 50 -sample random")
	fmt.Println("  go run . read-excel report.xlsx")
	fmt.Println("  go run . read-excel report.xlsx -sheet 2 -rows 30")
	fmt.Println()
	fmt.Println("  go run . process-data -input travel.xlsx \\")
	fmt.Println("    -columns \"country,risk_level\" \\")
	fmt.Println("    -prompt \"Extract destination country ISO code and assess risk level\"")
	fmt.Println()
	fmt.Println("Use '<command> -h' for help with a specific command")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "read-csv":
		err = tools.RunReadCSV(args)
	case "read-excel":
		err = tools.RunReadExcel(args)
	case "process-data":
		err = tools.RunProcessData(args)
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		fmt.Printf("Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}