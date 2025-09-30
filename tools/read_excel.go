package tools

import (
	"flag"
	"fmt"
	"strings"

	"ai-general-tool/common"
	"github.com/xuri/excelize/v2"
)

// RunReadExcel handles the read-excel command
func RunReadExcel(args []string) error {
	fs := flag.NewFlagSet("read-excel", flag.ExitOnError)

	// Define flags
	fileName := fs.String("file", "", "Excel file to read (required)")
	rowCount := fs.Int("rows", 20, "Number of rows to display")
	sampleType := fs.String("sample", "first", "Sample type: 'first' or 'random'")
	sheetIndex := fs.Int("sheet", 1, "Sheet number to read (1-based index)")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Handle positional argument for filename
	if *fileName == "" && fs.NArg() > 0 {
		*fileName = fs.Arg(0)
	}

	// Debug: print the values (commented out)
	// fmt.Printf("DEBUG: rows=%d, sample=%s, fileName=%s\n", *rowCount, *sampleType, *fileName)

	if *fileName == "" {
		fmt.Println("Error: Excel file name is required")
		fmt.Println("\nUsage:")
		fmt.Println("  read-excel <filename> [flags]")
		fmt.Println("  read-excel -file <filename> [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		return fmt.Errorf("missing required file argument")
	}

	// Open the Excel file
	f, err := excelize.OpenFile(*fileName)
	if err != nil {
		return fmt.Errorf("error opening file '%s': %v", *fileName, err)
	}
	defer f.Close()

	// Get sheet list
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return fmt.Errorf("no sheets found in Excel file")
	}

	// Validate sheet index
	if *sheetIndex < 1 || *sheetIndex > len(sheetList) {
		return fmt.Errorf("invalid sheet index %d. File has %d sheet(s)", *sheetIndex, len(sheetList))
	}

	// Get the sheet name
	sheetName := sheetList[*sheetIndex-1]

	// Read all rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("error reading sheet '%s': %v", sheetName, err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("sheet '%s' is empty", sheetName)
	}

	// Extract headers
	headers := rows[0]
	data := rows[1:]

	if len(data) == 0 {
		fmt.Println("Warning: Excel sheet contains only headers, no data rows")
		return nil
	}

	// Create sheet info string
	sheetInfo := fmt.Sprintf("Sheet %d of %d: \"%s\"", *sheetIndex, len(sheetList), sheetName)

	// Create data preview
	preview := &common.DataPreview{
		FileName:     *fileName,
		FileType:     "Excel Spreadsheet",
		SheetInfo:    sheetInfo,
		TotalRows:    len(data),
		TotalColumns: len(headers),
		Headers:      headers,
		SampleType:   *sampleType,
	}

	// Normalize data rows (ensure all rows have same number of columns)
	normalizedData := normalizeData(data, len(headers))

	// Analyze columns
	preview.Columns = analyzeExcelColumns(headers, normalizedData)

	// Select rows to display
	displayRows := selectExcelRows(normalizedData, *rowCount, *sampleType)
	preview.Rows = displayRows
	preview.RowsDisplayed = len(displayRows)

	// Display the preview
	displayExcelPreview(preview, len(sheetList))

	return nil
}

// normalizeData ensures all rows have the same number of columns
func normalizeData(data [][]string, colCount int) [][]string {
	normalized := make([][]string, len(data))
	for i, row := range data {
		normalized[i] = make([]string, colCount)
		for j := 0; j < colCount; j++ {
			if j < len(row) {
				normalized[i][j] = row[j]
			} else {
				normalized[i][j] = ""
			}
		}
	}
	return normalized
}

// analyzeExcelColumns analyzes the columns in the Excel data
func analyzeExcelColumns(headers []string, data [][]string) []common.ColumnInfo {
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

// selectExcelRows selects rows to display based on sample type
func selectExcelRows(data [][]string, count int, sampleType string) [][]string {
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

// displayExcelPreview displays the Excel data preview in formatted output
func displayExcelPreview(preview *common.DataPreview, totalSheets int) {
	separator := strings.Repeat("=", 80)

	// Header
	fmt.Println(separator)
	fmt.Printf("FILE: %s\n", preview.FileName)
	fmt.Printf("TYPE: %s (%s)\n", preview.FileType, preview.SheetInfo)
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

	rowNumberStart := 1
	if preview.SampleType == "random" {
		// For random sampling, we don't know the actual row numbers
		rowNumberStart = 1
	}

	for i, row := range preview.Rows {
		displayRow := append([]string{fmt.Sprintf("%d", rowNumberStart+i)}, row...)
		displayRows = append(displayRows, displayRow)
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
	fmt.Printf("\n[Showing %d of %d rows]\n", preview.RowsDisplayed, preview.TotalRows)
	fmt.Println()

	// Usage hints
	fmt.Println("USAGE HINTS:")
	fmt.Printf("• Use column index (0-%d) or column name to reference columns\n", len(preview.Headers)-1)
	fmt.Printf("• To see more rows: read-excel %s -rows 50\n", preview.FileName)
	if preview.SampleType == "random" {
		fmt.Printf("• To see first rows instead: read-excel %s -sample first\n", preview.FileName)
	} else {
		fmt.Printf("• To see random sample: read-excel %s -sample random\n", preview.FileName)
	}
	if totalSheets > 1 {
		fmt.Printf("• To select different sheet: read-excel %s -sheet 2\n", preview.FileName)
	}
	fmt.Println(separator)
}