# Claude Instructions for AI General Tool

## Purpose
This toolkit enables data enrichment through AI processing. Users will provide CSV or Excel files and describe what new columns they want generated using natural language. Your role is to help them explore their data, design prompts, and process their files.

## Available Tools

### process-data
Processes data files with AI to add enriched columns based on natural language instructions.

**When to use:** When user wants to add new columns to their data using AI processing.

**Command structure:**
```bash
go run . process-data [FLAGS] <filename>
```

**Important:** This tool requires an OpenAI API key in the .env file.

**Flags:**
- `-input <file>`: Input CSV or Excel file (required)
- `-output <file>`: Output file name (optional, defaults to input_enriched)
- `-columns <names>`: Comma-separated list of new column names to generate
- `-prompt <text>`: AI prompt describing what to extract/generate
- `-sample <n>`: Number of rows to test before full processing (default: 5)
- `-workers <n>`: Number of parallel workers (default: 10)
- `-batch-size <n>`: Save progress every N rows (default: 100)
- `-sheet <n>`: Excel sheet number, 1-based (default: 1)
- `-format <type>`: Output format: "same" or "csv" (default: same as input)

**Example usage patterns:**
```bash
# User: "Extract countries from travel descriptions"
go run . process-data \
  -input travel.xlsx \
  -columns "destination_country" \
  -prompt "Extract the destination country as ISO-3 code (e.g., AFG, FRA, USA)"

# User: "Categorize and prioritize items"
go run . process-data \
  -input items.csv \
  -columns "category,priority" \
  -prompt "Categorize the item and assign priority (LOW/MEDIUM/HIGH) based on description"

# User: "Translate and summarize"
go run . process-data \
  -input feedback.xlsx \
  -columns "translation,summary" \
  -prompt "Translate to English and provide a brief summary (max 20 words)"
```

### read-excel
Reads Excel files and displays comprehensive analysis.

**When to use:** When user mentions Excel files (.xlsx, .xls) or asks to explore spreadsheet data.

**Command structure:**
```bash
go run . read-excel [FLAGS] <filename>
```

**Important:** FLAGS MUST come BEFORE filename or they won't be parsed.

**Flags:**
- `-rows <n>`: Number of rows to display (default: 20)
- `-sample <type>`: Either "first" or "random" (default: "first")
- `-sheet <n>`: Sheet number, 1-based (default: 1)

**Example usage patterns:**
```bash
# User: "Show me what's in travel.xlsx"
go run . read-excel travel.xlsx

# User: "I want to see a random sample of the data"
go run . read-excel -rows 20 -sample random travel.xlsx

# User: "Show me just 5 rows to understand the structure"
go run . read-excel -rows 5 travel.xlsx

# User: "Check the second sheet"
go run . read-excel -sheet 2 travel.xlsx
```

### read-csv
Reads CSV files with similar functionality.

**When to use:** When user mentions CSV files or comma-separated data.

**Command structure:**
```bash
go run . read-csv [FLAGS] <filename>
```

**Flags:**
- `-rows <n>`: Number of rows to display (default: 20)
- `-sample <type>`: Either "first" or "random" (default: "first")
- `-delimiter <string>`: Field delimiter (default: ",")

**Example usage patterns:**
```bash
# User: "What's in data.csv?"
go run . read-csv data.csv

# User: "Show me random rows from the CSV"
go run . read-csv -rows 15 -sample random data.csv

# User: "It's a tab-separated file"
go run . read-csv -delimiter "\t" data.tsv
```

## Understanding the Output

The tools provide four sections:

1. **Summary Statistics**
   - Shows total rows/columns
   - Indicates sample type and size
   - Helps user understand data scale

2. **Column Analysis Table**
   - **Idx**: 0-based column index for referencing
   - **Type**: Detected data type (string/number/date/boolean/mixed/empty)
   - **Unique**: Count of distinct values (low = categorical, high = unique identifiers or free text)
   - **Nulls**: Missing data indicators
   - **Sample Values**: Examples from the column

3. **Data Preview**
   - Actual data rows in table format
   - Data types shown under headers
   - Row numbers for reference

4. **Usage Hints**
   - Suggestions for next commands

## Interaction Patterns

### Initial Data Exploration
When user provides a file without specific instructions:
```
User: "I have travel.xlsx"
Claude: Let me explore that file for you.
[Run: go run . read-excel -rows 10 travel.xlsx]
[Analyze output and explain what columns exist, data types, and potential enrichment opportunities]
```

### Guided Exploration
When user wants to understand their data better:
```
User: "I want to see what kind of descriptions are in my data"
Claude: Let me show you a random sample to see the variety of descriptions.
[Run: go run . read-excel -rows 20 -sample random travel.xlsx]
```

### Preparing for AI Processing
When user describes a transformation goal:
```
User: "I need to extract countries from these travel descriptions"
Claude: Let me first look at your data structure to understand what we're working with.
[Run: go run . read-excel -rows 10 travel.xlsx]
[Analyze the description column, note its index, check for patterns]

Based on the data, I can see:
- Column 1 contains travel descriptions
- There are 28,958 unique descriptions
- We'll create a new column with extracted country codes
```

## Key Insights to Provide

After running these tools, help users understand:

1. **Data Quality**
   - Point out high null counts
   - Identify mixed data types that might need cleaning
   - Note very high/low unique counts

2. **AI Processing Suitability**
   - High unique text values = good for extraction/summarization
   - Low unique values = might already be categorized
   - Date/number columns = might need different processing

3. **Next Steps**
   - Suggest which columns are good candidates for enrichment
   - Recommend sample size for testing prompts
   - Identify potential challenges (nulls, mixed types)

## Common Scenarios

### Scenario 1: Categorization Task
```
User: "I have products that need categories"
Action: go run . read-csv -rows 10 products.csv
Analysis: Identify product name column, check if categories exist, suggest AI categorization
```

### Scenario 2: Information Extraction
```
User: "Extract destinations from travel descriptions"
Action: go run . read-excel -rows 5 -sample random travel.xlsx
Analysis: Show variety in descriptions, confirm extraction is needed
```

### Scenario 3: Translation Task
```
User: "Translate survey responses"
Action: go run . read-csv -rows 10 survey.csv
Analysis: Identify text columns, check language variety, plan translation
```

## Important Notes

1. **File Location**: Assume files are in current directory unless user specifies path

2. **Flag Order**: Always put flags BEFORE filename:
   - ✅ `go run . read-excel -rows 10 file.xlsx`
   - ❌ `go run . read-excel file.xlsx -rows 10`

3. **Random Sampling**: Recommend `-sample random` for large files to better understand data variety

4. **Row Limits**: Start with 5-10 rows for quick exploration, use more for detailed analysis

5. **Column References**: Note column indices (0-based) for future processing

## Error Handling

- **File not found**: Ask user to confirm filename and location
- **No data rows**: File might only have headers
- **Parse errors**: Check delimiter for CSV files
- **Sheet not found**: Excel file might have fewer sheets than requested

## Integration with AI Processing

These tools are the first step in the pipeline:
1. **Explore** - Use read tools to understand structure
2. **Design** - Help user craft prompts based on data
3. **Test** - Process sample rows with AI
4. **Refine** - Iterate on prompt based on results
5. **Execute** - Run full processing with confirmed approach

## Process-Data Workflow

### Step 1: Data Exploration
First use read tools to understand the data:
```bash
go run . read-excel -rows 10 travel.xlsx
```

### Step 2: Design the Processing
Based on the data structure, help user craft:
- **Column names**: What new columns to create
- **Prompt**: Clear instructions for the AI

### Step 3: Test with Sample
The process-data tool automatically:
1. Tests on sample rows (default 5)
2. Shows results for review
3. Asks for confirmation before full processing

Example interaction:
```
User wants to extract countries from travel descriptions.

Claude: Let me first explore your data to understand the structure.
[Run: go run . read-excel -rows 5 travel.xlsx]

I can see you have travel descriptions in column 1. Let's test extracting countries:

[Run: go run . process-data -input travel.xlsx -columns "country_code" -prompt "Extract destination country as ISO-3 code" -sample 5]

The tool will show:
- Sample results for review
- Ask "Proceed with full processing? (y/n)"
```

### Step 4: Full Processing
When user confirms, the tool:
- Processes all rows with progress tracking
- Shows real-time statistics (rows completed, tokens used, estimated cost)
- Saves progress incrementally (every 100 rows or 30 seconds)
- Handles interruptions gracefully (Ctrl+C saves progress)

### Step 5: Output
The enriched file contains:
- All original columns
- New AI-generated columns appended
- Failed rows marked with "ERROR: <message>"

## Common Processing Patterns

### Single Column Extraction
```bash
go run . process-data \
  -input data.xlsx \
  -columns "extracted_value" \
  -prompt "Extract [specific information] from the description"
```

### Multiple Column Generation
```bash
go run . process-data \
  -input data.csv \
  -columns "field1,field2,field3" \
  -prompt "Extract field1 as X, field2 as Y, and field3 as Z"
```

### Categorization with Confidence
```bash
go run . process-data \
  -input items.xlsx \
  -columns "category,confidence" \
  -prompt "Categorize into (A/B/C) and provide confidence (HIGH/MEDIUM/LOW)"
```

### Translation and Processing
```bash
go run . process-data \
  -input multilingual.csv \
  -columns "english_text,sentiment,key_points" \
  -prompt "Translate to English, analyze sentiment (POSITIVE/NEUTRAL/NEGATIVE), extract 3 key points"
```

## Tips for Effective Prompts

1. **Be Specific**: Include examples in the prompt
   - Good: "Extract country as ISO-3 code (e.g., AFG, FRA, USA)"
   - Bad: "Get the country"

2. **Define Formats**: Specify exact output format
   - "Return date as YYYY-MM-DD"
   - "Use uppercase for categories"
   - "Limit summaries to 50 words"

3. **Handle Edge Cases**: Tell AI what to do when unsure
   - "If no date found, return 'N/A'"
   - "If multiple countries, return the first mentioned"
   - "If uncertain, return 'UNKNOWN'"

4. **Use Consistent Terminology**: Match the language in your data
   - If data uses "United States", don't ask for "USA" without clarification

## Performance Optimization

- **Workers**: Increase for faster processing (e.g., -workers 50)
- **Batch Size**: Larger batches = less frequent saves but more potential data loss on interrupt
- **Sample Size**: Test with more rows if data is highly varied

## Cost Considerations

The tool shows:
- Real-time token usage
- Estimated cost (based on GPT-4o-mini pricing)
- Final statistics after completion

For large datasets, consider:
- Testing thoroughly with samples first
- Using more specific prompts to reduce token usage
- Processing in batches if budget-constrained

Remember: The goal is to help users understand their data structure so they can effectively describe what new columns they want AI to generate.