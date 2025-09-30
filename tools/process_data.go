package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"


	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/xuri/excelize/v2"
)

// ProcessingTask represents a single row to process
type ProcessingTask struct {
	RowIndex int
	RowData  map[string]string // column name -> value
}

// ProcessingResult represents the result of processing a row
type ProcessingResult struct {
	RowIndex int
	RowData  map[string]string // original data
	Results  map[string]string // new column -> value
	Error    error
	Tokens   int
}

// ProcessingStats tracks overall progress
type ProcessingStats struct {
	TotalRows      int
	CompletedRows  int32
	FailedRows     int32
	TotalTokens    int64
	StartTime      time.Time
	EstimatedCost  float64
}

// RunProcessData handles the process-data command
func RunProcessData(args []string) error {
	fs := flag.NewFlagSet("process-data", flag.ExitOnError)

	// Define flags
	inputFile := fs.String("input", "", "Input file (CSV or Excel)")
	outputFile := fs.String("output", "", "Output file (optional, defaults to input_enriched)")
	columns := fs.String("columns", "", "Comma-separated list of new column names")
	prompt := fs.String("prompt", "", "AI prompt describing what to extract")
	sampleSize := fs.Int("sample", 5, "Number of rows to test before full processing")
	batchSize := fs.Int("batch-size", 100, "Save progress every N rows")
	workers := fs.Int("workers", 10, "Number of parallel workers")
	sheetIndex := fs.Int("sheet", 1, "Excel sheet number (1-based)")
	outputFormat := fs.String("format", "same", "Output format: same, csv")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Handle positional argument for filename
	if *inputFile == "" && fs.NArg() > 0 {
		*inputFile = fs.Arg(0)
	}

	// Validation
	if *inputFile == "" {
		return fmt.Errorf("input file is required")
	}
	if *columns == "" {
		return fmt.Errorf("columns to generate are required")
	}
	if *prompt == "" {
		return fmt.Errorf("AI prompt is required")
	}

	// Load API key
	if err := godotenv.Load(".env"); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY not found in environment")
	}

	// Initialize OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Parse column specifications
	columnSpecs := parseColumnSpecs(*columns)

	// Determine output file name
	if *outputFile == "" {
		ext := ".xlsx"
		if *outputFormat == "csv" || strings.HasSuffix(*inputFile, ".csv") {
			ext = ".csv"
		}
		base := strings.TrimSuffix(*inputFile, ".csv")
		base = strings.TrimSuffix(base, ".xlsx")
		*outputFile = base + "_enriched" + ext
	}

	// Load input data
	fmt.Printf("Loading %s...\n", *inputFile)
	headers, rows, err := loadInputFile(*inputFile, *sheetIndex)
	if err != nil {
		return fmt.Errorf("error loading input: %v", err)
	}

	fmt.Printf("Loaded %d rows with %d columns\n", len(rows), len(headers))

	// Test on sample first
	fmt.Println("\n=== TESTING ON SAMPLE ===")
	if err := testSample(&client, headers, rows, columnSpecs, *prompt, *sampleSize); err != nil {
		return fmt.Errorf("sample test failed: %v", err)
	}

	// Ask for confirmation
	fmt.Print("\nProceed with full processing? (y/n): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		fmt.Println("Processing cancelled.")
		return nil
	}

	// Process full dataset
	fmt.Println("\n=== PROCESSING FULL DATASET ===")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nInterrupt received. Saving progress...")
		cancel()
	}()

	// Process data
	enrichedRows, stats := processFullDataset(
		ctx,
		&client,
		headers,
		rows,
		columnSpecs,
		*prompt,
		*workers,
		*batchSize,
		*outputFile,
	)

	// Save final output
	fmt.Println("\nSaving final output...")
	if err := saveOutputFile(*outputFile, headers, enrichedRows, columnSpecs, *outputFormat); err != nil {
		return fmt.Errorf("error saving output: %v", err)
	}

	// Print final statistics
	printFinalStats(stats)
	fmt.Printf("\nOutput saved to: %s\n", *outputFile)

	return nil
}

// parseColumnSpecs parses column specifications (with optional type hints)
func parseColumnSpecs(columnsStr string) []ColumnSpec {
	parts := strings.Split(columnsStr, ",")
	specs := make([]ColumnSpec, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, ":") {
			// Has type hint
			subparts := strings.SplitN(part, ":", 2)
			specs[i] = ColumnSpec{
				Name:     strings.TrimSpace(subparts[0]),
				DataType: strings.TrimSpace(subparts[1]),
			}
		} else {
			// Default to string
			specs[i] = ColumnSpec{
				Name:     part,
				DataType: "string",
			}
		}
	}

	return specs
}

// ColumnSpec represents a column specification
type ColumnSpec struct {
	Name     string
	DataType string
}

// loadInputFile loads data from CSV or Excel
func loadInputFile(filename string, sheetIndex int) ([]string, [][]string, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return loadCSV(filename)
	}
	return loadExcel(filename, sheetIndex)
}

// loadCSV loads data from a CSV file
func loadCSV(filename string) ([]string, [][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	allData, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(allData) < 2 {
		return nil, nil, fmt.Errorf("file must have headers and at least one data row")
	}

	return allData[0], allData[1:], nil
}

// loadExcel loads data from an Excel file
func loadExcel(filename string, sheetIndex int) ([]string, [][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if sheetIndex < 1 || sheetIndex > len(sheets) {
		return nil, nil, fmt.Errorf("invalid sheet index %d (file has %d sheets)", sheetIndex, len(sheets))
	}

	sheetName := sheets[sheetIndex-1]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("sheet must have headers and at least one data row")
	}

	return rows[0], rows[1:], nil
}

// testSample tests processing on a small sample
func testSample(client *openai.Client, headers []string, rows [][]string, columnSpecs []ColumnSpec, userPrompt string, sampleSize int) error {
	fmt.Printf("Testing on %d sample rows...\n\n", sampleSize)

	// Take sample rows
	sample := rows
	if len(rows) > sampleSize {
		sample = rows[:sampleSize]
	}

	// Process each sample row
	for i, row := range sample {
		rowData := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				rowData[header] = row[j]
			} else {
				rowData[header] = ""
			}
		}

		result, err := processRow(context.Background(), client, rowData, columnSpecs, userPrompt)
		if err != nil {
			fmt.Printf("Row %d: ERROR - %v\n", i+1, err)
			continue
		}

		fmt.Printf("Row %d:\n", i+1)
		fmt.Printf("  Input: %v\n", truncateMap(rowData, 50))
		fmt.Printf("  Output: %v\n", result.Results)
	}

	return nil
}

// processRow processes a single row using OpenAI
func processRow(ctx context.Context, client *openai.Client, rowData map[string]string, columnSpecs []ColumnSpec, userPrompt string) (*ProcessingResult, error) {
	// Build the context for the AI
	var dataContext strings.Builder
	for key, value := range rowData {
		if value == "" {
			dataContext.WriteString(fmt.Sprintf("%s: [empty]\n", key))
		} else {
			dataContext.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}

	// Build JSON schema for structured output
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, spec := range columnSpecs {
		properties[spec.Name] = map[string]interface{}{
			"type":        "string", // For now, all strings
			"description": fmt.Sprintf("Value for %s column", spec.Name),
		}
		required = append(required, spec.Name)
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}

	// System prompt
	systemPrompt := `You are a data processing assistant. You analyze input data and extract or generate the requested information in a structured format.
Always return valid values for all requested fields. If a value cannot be determined, use "N/A" or an appropriate default.
Be consistent in your formatting across all rows.`

	// User message combining data and prompt
	userMessage := fmt.Sprintf("Data:\n%s\n\nTask: %s", dataContext.String(), userPrompt)

	// Call OpenAI with function calling for structured output
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userMessage),
		},
		Functions: []openai.ChatCompletionNewParamsFunction{
			{
				Name:        "extract_data",
				Description: openai.String("Extract or generate the requested data fields"),
				Parameters:  openai.FunctionParameters(schema),
			},
		},
		Temperature: openai.Float(0.3),
		MaxTokens:   openai.Int(500),
	}

	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	choice := completion.Choices[0]
	if choice.Message.FunctionCall.Name == "" {
		return nil, fmt.Errorf("no function call in response")
	}

	// Parse the function arguments
	var results map[string]string
	if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &results); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	tokens := 0
	if completion.Usage.TotalTokens > 0 {
		tokens = int(completion.Usage.TotalTokens)
	}

	return &ProcessingResult{
		Results: results,
		Tokens:  tokens,
	}, nil
}

// processFullDataset processes the entire dataset
func processFullDataset(
	ctx context.Context,
	client *openai.Client,
	headers []string,
	rows [][]string,
	columnSpecs []ColumnSpec,
	userPrompt string,
	workerCount int,
	batchSize int,
	outputFile string,
) ([][]string, *ProcessingStats) {

	stats := &ProcessingStats{
		TotalRows: len(rows),
		StartTime: time.Now(),
	}

	// Create channels
	taskChan := make(chan ProcessingTask, workerCount*2)
	resultChan := make(chan ProcessingResult, workerCount*2)

	// Create enriched rows (copy of original with space for new columns)
	enrichedRows := make([][]string, len(rows))
	for i, row := range rows {
		enrichedRows[i] = make([]string, len(row)+len(columnSpecs))
		copy(enrichedRows[i], row)
	}

	// Mutex for protecting enrichedRows
	var rowMutex sync.Mutex

	// Start result collector
	doneChan := make(chan bool)
	go collectResults(ctx, resultChan, enrichedRows, headers, columnSpecs, &rowMutex, stats, batchSize, outputFile, doneChan)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go processWorker(ctx, client, headers, columnSpecs, userPrompt, taskChan, resultChan, &wg, stats)
	}

	// Send tasks
	go func() {
		for i, row := range rows {
			rowData := make(map[string]string)
			for j, header := range headers {
				if j < len(row) {
					rowData[header] = row[j]
				} else {
					rowData[header] = ""
				}
			}

			select {
			case <-ctx.Done():
				break
			case taskChan <- ProcessingTask{RowIndex: i, RowData: rowData}:
			}
		}
		close(taskChan)
	}()

	// Wait for workers to finish
	wg.Wait()
	close(resultChan)
	<-doneChan

	return enrichedRows, stats
}

// processWorker is a worker goroutine
func processWorker(
	ctx context.Context,
	client *openai.Client,
	headers []string,
	columnSpecs []ColumnSpec,
	userPrompt string,
	taskChan <-chan ProcessingTask,
	resultChan chan<- ProcessingResult,
	wg *sync.WaitGroup,
	stats *ProcessingStats,
) {
	defer wg.Done()

	for task := range taskChan {
		select {
		case <-ctx.Done():
			return
		default:
			result, err := processRow(ctx, client, task.RowData, columnSpecs, userPrompt)

			processingResult := ProcessingResult{
				RowIndex: task.RowIndex,
				RowData:  task.RowData,
			}

			if err != nil {
				processingResult.Error = err
				// Put error message in results
				processingResult.Results = make(map[string]string)
				for _, spec := range columnSpecs {
					processingResult.Results[spec.Name] = fmt.Sprintf("ERROR: %v", err)
				}
			} else {
				processingResult.Results = result.Results
				processingResult.Tokens = result.Tokens
			}

			resultChan <- processingResult
		}
	}
}

// collectResults collects and saves results
func collectResults(
	ctx context.Context,
	resultChan <-chan ProcessingResult,
	enrichedRows [][]string,
	headers []string,
	columnSpecs []ColumnSpec,
	rowMutex *sync.Mutex,
	stats *ProcessingStats,
	batchSize int,
	outputFile string,
	doneChan chan<- bool,
) {
	saveTimer := time.NewTicker(30 * time.Second)
	defer saveTimer.Stop()

	processedCount := 0

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				doneChan <- true
				return
			}

			// Update enriched rows
			rowMutex.Lock()
			row := enrichedRows[result.RowIndex]
			startIdx := len(headers)
			for i, spec := range columnSpecs {
				if val, ok := result.Results[spec.Name]; ok {
					row[startIdx+i] = val
				} else {
					row[startIdx+i] = ""
				}
			}
			rowMutex.Unlock()

			// Update stats
			if result.Error == nil {
				atomic.AddInt32(&stats.CompletedRows, 1)
				atomic.AddInt64(&stats.TotalTokens, int64(result.Tokens))
			} else {
				atomic.AddInt32(&stats.FailedRows, 1)
			}

			processedCount++
			printProgress(stats)

			// Save periodically
			if processedCount%batchSize == 0 {
				saveProgress(outputFile, headers, enrichedRows, columnSpecs, rowMutex)
			}

		case <-saveTimer.C:
			// Periodic save
			saveProgress(outputFile, headers, enrichedRows, columnSpecs, rowMutex)

		case <-ctx.Done():
			// Save on interrupt
			saveProgress(outputFile, headers, enrichedRows, columnSpecs, rowMutex)
			doneChan <- true
			return
		}
	}
}

// saveProgress saves current progress to temp file
func saveProgress(outputFile string, headers []string, enrichedRows [][]string, columnSpecs []ColumnSpec, rowMutex *sync.Mutex) {
	tempFile := outputFile + ".tmp"

	rowMutex.Lock()
	defer rowMutex.Unlock()

	// Build full headers
	fullHeaders := append(headers, getColumnNames(columnSpecs)...)

	if strings.HasSuffix(outputFile, ".csv") {
		saveCSV(tempFile, fullHeaders, enrichedRows)
	} else {
		saveExcel(tempFile, fullHeaders, enrichedRows)
	}
}

// saveOutputFile saves the final output
func saveOutputFile(outputFile string, headers []string, enrichedRows [][]string, columnSpecs []ColumnSpec, format string) error {
	// Build full headers
	fullHeaders := append(headers, getColumnNames(columnSpecs)...)

	if format == "csv" || strings.HasSuffix(outputFile, ".csv") {
		return saveCSV(outputFile, fullHeaders, enrichedRows)
	}
	return saveExcel(outputFile, fullHeaders, enrichedRows)
}

// saveCSV saves data to CSV
func saveCSV(filename string, headers []string, rows [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// saveExcel saves data to Excel
func saveExcel(filename string, headers []string, rows [][]string) error {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", columnIndexToLetter(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for i, row := range rows {
		for j, value := range row {
			cell := fmt.Sprintf("%s%d", columnIndexToLetter(j), i+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	return f.SaveAs(filename)
}

// Helper functions

func getColumnNames(specs []ColumnSpec) []string {
	names := make([]string, len(specs))
	for i, spec := range specs {
		names[i] = spec.Name
	}
	return names
}

func columnIndexToLetter(index int) string {
	result := ""
	for index >= 0 {
		result = string('A'+index%26) + result
		index = index/26 - 1
	}
	return result
}

func truncateMap(m map[string]string, maxLen int) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if len(v) > maxLen {
			result[k] = v[:maxLen] + "..."
		} else {
			result[k] = v
		}
	}
	return result
}

func printProgress(stats *ProcessingStats) {
	completed := atomic.LoadInt32(&stats.CompletedRows)
	failed := atomic.LoadInt32(&stats.FailedRows)
	total := stats.TotalRows
	tokens := atomic.LoadInt64(&stats.TotalTokens)

	percentage := float64(completed+failed) * 100 / float64(total)
	elapsed := time.Since(stats.StartTime)

	// Estimate cost (GPT-4o-mini pricing)
	costPerMillion := 0.15  // $0.15 per 1M input tokens
	costPer1MOutput := 0.60 // $0.60 per 1M output tokens
	estimatedCost := float64(tokens) / 1000000 * ((costPerMillion + costPer1MOutput) / 2)

	fmt.Printf("\rProgress: %d/%d (%.1f%%) | Failed: %d | Tokens: %d | Cost: $%.4f | Elapsed: %s",
		completed, total, percentage, failed, tokens, estimatedCost, elapsed.Round(time.Second))
}

func printFinalStats(stats *ProcessingStats) {
	fmt.Println("\n\n=== FINAL STATISTICS ===")
	fmt.Printf("Total rows processed: %d\n", stats.CompletedRows+stats.FailedRows)
	fmt.Printf("Successful: %d\n", stats.CompletedRows)
	fmt.Printf("Failed: %d\n", stats.FailedRows)
	fmt.Printf("Total tokens used: %d\n", stats.TotalTokens)

	// Calculate final cost
	costPerMillion := 0.15
	costPer1MOutput := 0.60
	estimatedCost := float64(stats.TotalTokens) / 1000000 * ((costPerMillion + costPer1MOutput) / 2)
	fmt.Printf("Estimated cost: $%.4f\n", estimatedCost)

	elapsed := time.Since(stats.StartTime)
	fmt.Printf("Total time: %s\n", elapsed.Round(time.Second))

	if stats.CompletedRows > 0 {
		avgTime := elapsed / time.Duration(stats.CompletedRows)
		fmt.Printf("Average time per row: %s\n", avgTime.Round(time.Millisecond))
	}
}