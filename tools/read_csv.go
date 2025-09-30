package tools

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	"ai-general-tool/common"
)

// RunReadCSV handles the read-csv command
func RunReadCSV(args []string) error {
	fs := flag.NewFlagSet("read-csv", flag.ExitOnError)

	// Define flags
	fileName := fs.String("file", "", "CSV file to read (required)")
	rowCount := fs.Int("rows", 20, "Number of rows to display")
	sampleType := fs.String("sample", "first", "Sample type: 'first' or 'random'")
	delimiter := fs.String("delimiter", ",", "CSV delimiter")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Handle positional argument for filename
	if *fileName == "" && fs.NArg() > 0 {
		*fileName = fs.Arg(0)
	}

	if *fileName == "" {
		fmt.Println("Error: CSV file name is required")
		fmt.Println("\nUsage:")
		fmt.Println("  read-csv <filename> [flags]")
		fmt.Println("  read-csv -file <filename> [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		return fmt.Errorf("missing required file argument")
	}

	// Open the CSV file
	file, err := os.Open(*fileName)
	if err != nil {
		return fmt.Errorf("error opening file '%s': %v", *fileName, err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	reader.Comma = []rune(*delimiter)[0]
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read all data (for analysis)
	allData, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %v", err)
	}

	if len(allData) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Extract headers
	headers := allData[0]
	data := allData[1:]

	if len(data) == 0 {
		fmt.Println("Warning: CSV file contains only headers, no data rows")
		return nil
	}

	// Create data preview
	preview := &common.DataPreview{
		FileName:     *fileName,
		FileType:     "CSV File",
		TotalRows:    len(data),
		TotalColumns: len(headers),
		Headers:      headers,
		SampleType:   *sampleType,
	}

	// Analyze columns
	preview.Columns = analyzeColumns(headers, data)

	// Select rows to display
	displayRows := selectRows(data, *rowCount, *sampleType)
	preview.Rows = displayRows
	preview.RowsDisplayed = len(displayRows)

	// Display the preview
	displayPreview(preview)

	return nil
}

// analyzeColumns analyzes the columns in the data
func analyzeColumns(headers []string, data [][]string) []common.ColumnInfo {
	columns := make([]common.ColumnInfo, len(headers))

	for i, header := range headers {
		// Collect all values for this column
		var values []string
		for _, row := range data {
			if i < len(row) {
				values = append(values, row[i])
			} else {
				values = append(values, "")
			}
		}

		// Get unique values
		uniqueValues := common.GetUniqueValues(values)

		// Get sample values (first 5 unique)
		sampleValues := uniqueValues
		if len(sampleValues) > 5 {
			sampleValues = sampleValues[:5]
		}

		// Truncate sample values for display
		for j := range sampleValues {
			sampleValues[j] = common.TruncateString(sampleValues[j], 15)
		}

		columns[i] = common.ColumnInfo{
			Index:        i,
			Name:         header,
			DataType:     common.DetectDataType(values),
			UniqueCount:  len(uniqueValues),
			NullCount:    common.CountNulls(values),
			TotalCount:   len(values),
			SampleValues: sampleValues,
		}
	}

	return columns
}

// selectRows selects rows to display based on sample type
func selectRows(data [][]string, count int, sampleType string) [][]string {
	if len(data) <= count {
		return data
	}

	if sampleType == "random" {
		indices := common.GenerateRandomIndices(count, len(data))
		result := make([][]string, len(indices))
		for i, idx := range indices {
			result[i] = data[idx]
		}
		return result
	}

	// Default to first rows
	return data[:count]
}

// displayPreview displays the data preview in formatted output
func displayPreview(preview *common.DataPreview) {
	separator := strings.Repeat("=", 80)

	// Header
	fmt.Println(separator)
	fmt.Printf("FILE: %s\n", preview.FileName)
	fmt.Printf("TYPE: %s\n", preview.FileType)
	fmt.Println(separator)
	fmt.Println()

	// Summary Statistics
	fmt.Println("SUMMARY STATISTICS:")
	fmt.Printf("Total Rows: %d\n", preview.TotalRows)
	fmt.Printf("Total Columns: %d\n", preview.TotalColumns)
	fmt.Printf("Rows Displayed: %d (%s)\n", preview.RowsDisplayed, preview.SampleType)
	fmt.Println()

	// Column Analysis
	fmt.Println("COLUMN ANALYSIS:")
	analysisHeaders := []string{"Idx", "Column Name", "Type", "Unique", "Nulls", "Sample Values"}
	var analysisRows [][]string

	for _, col := range preview.Columns {
		nullPercent := common.FormatPercentage(col.NullCount, col.TotalCount)
		sampleStr := strings.Join(col.SampleValues, ", ")
		if len(col.SampleValues) < col.UniqueCount {
			sampleStr += "..."
		}

		row := []string{
			fmt.Sprintf("%d", col.Index),
			common.TruncateString(col.Name, 20),
			string(col.DataType),
			fmt.Sprintf("%d", col.UniqueCount),
			fmt.Sprintf("%d (%s)", col.NullCount, nullPercent),
			sampleStr,
		}
		analysisRows = append(analysisRows, row)
	}

	fmt.Println(common.FormatTable(analysisHeaders, analysisRows, 120))
	fmt.Println()

	// Data Preview
	if preview.SampleType == "random" {
		fmt.Println("DATA PREVIEW (Random Sample):")
	} else {
		fmt.Println("DATA PREVIEW:")
	}

	// Add row numbers to the display
	displayHeaders := append([]string{"Row"}, preview.Headers...)

	// Add data type row
	typeRow := []string{""}
	for _, col := range preview.Columns {
		typeRow = append(typeRow, fmt.Sprintf("[%s]", col.DataType))
	}

	// Create display rows with row numbers
	var displayRows [][]string
	displayRows = append(displayRows, typeRow) // Add type row

	for i, row := range preview.Rows {
		displayRow := append([]string{fmt.Sprintf("%d", i+1)}, row...)
		displayRows = append(displayRows, displayRow)

		// Limit display to avoid overwhelming output
		if i >= 19 {
			break
		}
	}

	// Add ellipsis row if there are more
	if preview.TotalRows > preview.RowsDisplayed {
		ellipsisRow := []string{"..."}
		for range preview.Headers {
			ellipsisRow = append(ellipsisRow, "...")
		}
		displayRows = append(displayRows, ellipsisRow)
	}

	fmt.Println(common.FormatTable(displayHeaders, displayRows, 150))
	fmt.Printf("\n[Showing %d of %d rows]\n", common.Min(preview.RowsDisplayed, 20), preview.TotalRows)
	fmt.Println()

	// Usage hints
	fmt.Println("USAGE HINTS:")
	fmt.Printf("• Use column index (0-%d) or column name to reference columns\n", len(preview.Headers)-1)
	fmt.Printf("• To see more rows: read-csv %s -rows 50\n", preview.FileName)
	if preview.SampleType == "random" {
		fmt.Printf("• To see first rows instead: read-csv %s -sample first\n", preview.FileName)
	} else {
		fmt.Printf("• To see random sample: read-csv %s -sample random\n", preview.FileName)
	}
	fmt.Println(separator)
}