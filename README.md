# AI General Tool - Data Enrichment Through Natural Language

Transform your CSV and Excel data with AI-powered column generation using simple, natural language descriptions.

## Overview

AI General Tool is a command-line utility that empowers anyone to enrich their tabular data without writing code. Simply describe what new columns you want in plain English, and the tool uses AI to generate them automatically.

**Perfect for:**
- Data analysts needing quick transformations
- Business users without programming skills
- Anyone working with CSV or Excel files who needs AI-powered data enrichment

**Key Benefits:**
- 🚀 No coding required - use natural language
- 📊 Works with CSV and Excel files
- ⚡ Parallel processing for large datasets
- 💰 Cost estimation before processing
- 🔄 Incremental progress saves
- ✅ Test-first approach prevents costly mistakes

## Prerequisites

- Go 1.24 or higher
- Anthropic API key
- CSV or Excel files to process

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/ai-general-tool.git
   cd ai-general-tool
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up your Anthropic API key:
   ```bash
   # Copy the example environment file
   cp .env.example .env

   # Edit .env and replace with your actual API key
   # Get your key from: https://platform.claude.com/settings/keys
   ```

4. Test the installation:
   ```bash
   go run . --help
   ```

## Quick Start

### Step 1: Explore Your Data
```bash
# View first 10 rows of an Excel file
go run . read-excel travel.xlsx -rows 10

# See random sample from CSV
go run . read-csv data.csv -sample random -rows 20
```

### Step 2: Test AI Enrichment
```bash
# Extract countries from travel descriptions (tests on 5 rows first)
go run . process-data \
  -input travel.xlsx \
  -columns "country_code" \
  -prompt "Extract destination country as ISO-3 code (e.g., USA, FRA, CHN)"
```

### Step 3: Process Full Dataset
When prompted after the test, type `y` to process all rows with progress tracking.

## Command Reference

### `read-csv` - Analyze CSV Files

Displays comprehensive analysis of CSV files including column types, unique values, nulls, and data preview.

**Usage:**
```bash
go run . read-csv [FLAGS] <filename>
```

**Flags:**
- `-rows <n>`: Number of rows to display (default: 20)
- `-sample <type>`: "first" or "random" (default: "first")
- `-delimiter <char>`: Field delimiter (default: ",")

**Examples:**
```bash
# Basic usage
go run . read-csv data.csv

# Random sample of 30 rows
go run . read-csv -rows 30 -sample random data.csv

# Tab-separated file
go run . read-csv -delimiter "\t" data.tsv
```

### `read-excel` - Analyze Excel Files

Provides detailed analysis of Excel files with multi-sheet support.

**Usage:**
```bash
go run . read-excel [FLAGS] <filename>
```

**Flags:**
- `-rows <n>`: Number of rows to display (default: 20)
- `-sample <type>`: "first" or "random" (default: "first")
- `-sheet <n>`: Sheet number, 1-based (default: 1)

**Examples:**
```bash
# View first sheet
go run . read-excel report.xlsx

# Check second sheet with random sampling
go run . read-excel -sheet 2 -sample random report.xlsx

# Quick preview with just 5 rows
go run . read-excel -rows 5 report.xlsx
```

### `process-data` - AI-Powered Data Enrichment

Processes data files with AI to add new columns based on natural language instructions.

**Usage:**
```bash
go run . process-data [FLAGS]
```

**Required Flags:**
- `-input <file>`: Input CSV or Excel file
- `-columns <names>`: Comma-separated list of new column names
- `-prompt <text>`: Natural language description of what to generate

**Optional Flags:**
- `-output <file>`: Output filename (default: input_enriched)
- `-sample <n>`: Rows to test before full processing (default: 5)
- `-workers <n>`: Parallel workers for speed (default: 10, max: 100)
- `-batch-size <n>`: Save progress every N rows (default: 100)
- `-sheet <n>`: Excel sheet number, 1-based (default: 1)
- `-format <type>`: Output format: "same" or "csv" (default: same as input)

**Examples:**
```bash
# Extract single value
go run . process-data \
  -input customers.csv \
  -columns "country" \
  -prompt "Extract the country name from the address field"

# Generate multiple columns
go run . process-data \
  -input feedback.xlsx \
  -columns "sentiment,category,priority" \
  -prompt "Analyze sentiment (POSITIVE/NEUTRAL/NEGATIVE), categorize feedback type, assign priority (HIGH/MEDIUM/LOW)"

# Fast processing with more workers
go run . process-data \
  -input large_dataset.csv \
  -columns "summary" \
  -prompt "Summarize in 10 words or less" \
  -workers 50
```

## Use Cases & Examples

### 1. Travel & Security
**Task:** Extract destinations and assess risk levels
```bash
go run . process-data \
  -input travel_requests.xlsx \
  -columns "destination_country,risk_level,requires_approval" \
  -prompt "Extract country as ISO-3 code, assess risk (LOW/MEDIUM/HIGH/CRITICAL), mark TRUE if requires approval (HIGH or CRITICAL)"
```

### 2. Human Resources
**Task:** Parse experience and skills from resumes
```bash
go run . process-data \
  -input job_applications.csv \
  -columns "years_experience,primary_skills,seniority_level" \
  -prompt "Extract total years of experience as number, list top 3 technical skills, classify seniority (JUNIOR/MID/SENIOR/LEAD)"
```

### 3. Customer Support
**Task:** Translate and analyze customer feedback
```bash
go run . process-data \
  -input multilingual_feedback.xlsx \
  -columns "english_text,sentiment,action_required" \
  -prompt "Translate to English, determine sentiment, mark TRUE if requires immediate action"
```

### 4. Finance & Procurement
**Task:** Categorize and validate expenses
```bash
go run . process-data \
  -input expenses.csv \
  -columns "category,budget_code,compliance_flag" \
  -prompt "Categorize expense (TRAVEL/SUPPLIES/SERVICES/OTHER), assign budget code, flag any compliance issues"
```

### 5. Data Quality
**Task:** Clean and standardize data
```bash
go run . process-data \
  -input contacts.xlsx \
  -columns "clean_phone,country_code,valid_email" \
  -prompt "Format phone to international standard, extract country code, validate email (TRUE/FALSE)"
```

## Writing Effective Prompts

### Best Practices

1. **Be Specific with Formats**
   ```
   Good: "Extract date as YYYY-MM-DD format"
   Bad:  "Get the date"
   ```

2. **Provide Examples**
   ```
   Good: "Extract country as ISO-3 code (e.g., USA, GBR, FRA)"
   Bad:  "Extract country"
   ```

3. **Handle Edge Cases**
   ```
   Good: "If no date found, return 'N/A'. If multiple dates, use the earliest"
   Bad:  "Extract the date"
   ```

4. **Use Consistent Terms**
   ```
   Good: "Categorize as FOOD/TRANSPORT/LODGING/OTHER"
   Bad:  "Categorize the expense"
   ```

5. **Limit Output Length**
   ```
   Good: "Summarize in maximum 20 words"
   Bad:  "Provide a summary"
   ```

### Prompt Templates

**Extraction:**
```
"Extract [field] as [format]. If not found, return 'N/A'. Examples: [examples]"
```

**Categorization:**
```
"Categorize into one of: [CATEGORY1/CATEGORY2/CATEGORY3]. If uncertain, use 'OTHER'"
```

**Translation:**
```
"Translate to [language]. Preserve numbers and proper nouns as-is"
```

**Analysis:**
```
"Analyze [aspect] and classify as [OPTION1/OPTION2/OPTION3] based on [criteria]"
```

## Performance & Cost Optimization

### Speed Optimization
- **Increase workers** for faster processing: `-workers 50`
- **Larger batch sizes** reduce save frequency: `-batch-size 500`
- **Optimize prompts** to be concise and specific

### Cost Management
- **Token usage** is displayed in real-time during processing
- **Estimated costs** shown based on Claude Sonnet 5.5 pricing
- **Test first** with small samples to refine prompts
- **Shorter prompts** reduce token usage

### Processing Large Datasets
```bash
# For datasets with 100k+ rows
go run . process-data \
  -input huge_file.csv \
  -columns "result" \
  -prompt "Process efficiently" \
  -workers 100 \
  -batch-size 1000
```

## Data Format Requirements

### Input Files
- **First row must contain headers**
- **Supported formats:** CSV, Excel (.xlsx, .xls)
- **Character encoding:** UTF-8 recommended
- **File size:** No hard limit, but larger files take longer

### Output Format
- Original columns are preserved
- New AI-generated columns are appended
- Failed rows show "ERROR: <message>" in new columns
- Progress is saved incrementally

## Error Handling & Recovery

### Automatic Recovery
- Progress saves every batch (default: 100 rows)
- Interruption with Ctrl+C saves current progress
- Progress is saved, but automatic restart/resume is not implemented; do not rerun completed rows blindly.

### Common Issues

**Missing API Key:**
```
Error: ANTHROPIC_API_KEY not found in environment
Solution: Add key to .env file
```

**Invalid Sheet Number:**
```
Error: Sheet 3 not found (file has 2 sheets)
Solution: Use -sheet 1 or -sheet 2
```

**Rate Limiting:**
```
Error: API rate limit exceeded
Solution: Reduce -workers parameter or wait
```

## Advanced Configuration

### Environment Variables
```bash
# Required
ANTHROPIC_API_KEY=your_key_here

# Optional (future features)
# Default model: claude-sonnet-5-5 (override with -model)
# Model selection: use -model claude-sonnet-5-5
```

### Default Values
- Sample size: 5 rows
- Workers: 10 parallel processors
- Batch size: 100 rows
- Output format: Same as input

## Architecture

The tool follows a modular architecture:

```
ai-general-tool/
├── main.go           # CLI dispatcher
├── tools/
│   ├── read_csv.go     # CSV analysis
│   ├── read_excel.go   # Excel analysis
│   └── process_data.go # AI processing
├── common/
│   ├── types.go        # Shared types
│   └── utils.go        # Utility functions
└── .env               # Configuration
```

Each tool is self-contained and can be tested independently. The system uses:
- **Parallel processing** for scalability
- **Streaming I/O** for memory efficiency
- **Incremental saves** for reliability
- **Token tracking** for cost management

## Contributing

We welcome contributions! Please:

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Code Style
- Follow Go standard formatting (`go fmt`)
- Add comments for exported functions
- Keep functions focused and testable
- Handle errors explicitly

## Troubleshooting

### Installation Issues
- Ensure Go version 1.24+ with `go version`
- Run `go mod tidy` if dependencies fail
- Check file permissions for reading/writing

### Processing Issues
- Verify API key is valid and has credits
- Start with small samples to test prompts
- Check input file has headers in first row
- Ensure output directory is writable

### Performance Issues
- Reduce worker count if hitting rate limits
- Increase batch size for better efficiency
- Optimize prompts to reduce token usage
- Process during off-peak hours for better API response

## Support

- **Issues:** Report bugs via GitHub Issues
- **Documentation:** See CLAUDE.md for detailed instructions
- **Examples:** Check the examples folder for more use cases

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with Anthropic Claude
- Excel handling via excelize library
- Inspired by the need for accessible AI data processing

---

**Made with ❤️ for data professionals who want AI power without the complexity**
## CPD direction classification

The default provider is Anthropic Messages with `claude-sonnet-5-5` and JSON structured outputs. It uses the standard Go HTTP client with bounded workers, request timeouts and retries for transient failures. No temperature or forced tool-call parameters are sent.

Prepare one row per indicator definition/context and preserve its stable indicator ID. Avoid feeding repeated yearly measurements, and keep components separate when they measure different concepts.

```bash
go run . process-data -input cpd-pilot.csv -output cpd-directions.csv \
  -preset cpd-direction -limit 100 -workers 5 -sample 0 -yes
```

The preset returns `direction`, `response_kind`, and `explanation`. Direction is one of `HIGHER_IS_BETTER`, `LOWER_IS_BETTER`, `NOT_APPLICABLE`, or `UNCLEAR`. It does not infer the desired direction from whether an actual is above/below its target. Preserve uncertain and categorical cases for review. Inputs are data rather than instructions.

`-limit` restricts the input to the first N rows; prepare a representative/shuffled pilot file first. `-sample` tests rows before the confirmation and incurs separate API calls; `-sample 0` avoids duplicate sample calls when running an already prepared pilot. `-yes` skips confirmation for the selected rows only. Without it the tool retains its interactive confirmation.

The output sidecar `.run.json` records the model, prompt and input hashes, token counts and unreviewed status. It never contains the API key. Keep the output and sidecar together. Costs use Sonnet 5.5 standard pricing ($2/M input and $10/M output); the estimate is not valid if another model is selected and excludes sample calls and failed responses. Progress saves do not implement automatic resume.

Spot-check a manually labelled pilot, inspect uncertain cases, and review the resulting mapping before using it in production scoring. This command never updates Performance App data.
