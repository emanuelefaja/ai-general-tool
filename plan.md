# AI-Powered General Purpose Data Enrichment Toolkit

## Core Concept

A flexible system where users can take their existing tabular data (CSV/Excel) and use natural language to describe what new columns they want AI to generate - without needing to write code themselves.

## Key Requirements

### 1. Universal Data Input
- Handle both CSV and Excel files
- Treat first row as headers
- Support multiple sheets (for Excel)
- Auto-detect data types and structure

### 2. Natural Language Interface
- Users describe transformations in plain English via Claude Code
- No coding knowledge required
- System interprets intent and generates appropriate prompts

### 3. Flexible AI Processing
- Support any transformation type:
  - Categorization
  - Translation
  - Summarization
  - Extraction
  - Classification
  - Sentiment analysis
  - Content generation
  - Data validation
  - And more...

### 4. Iterative Refinement Workflow
- Test on small sample (e.g., 5 rows)
- Show results to user for review
- Get feedback and adjust prompts/approach
- Iterate until output meets requirements
- Once satisfied, scale up with parallel processing (e.g., 150 workers)

## Architecture Pattern (from Demo-Builder)

- Individual Go tools for each operation:
  - `read-data` - Load CSV/Excel files
  - `preview-transform` - Test transformation on sample rows
  - `apply-transform` - Run full transformation with parallel processing
  - `export-results` - Save enriched data
- Common utilities for:
  - File handling (CSV/Excel I/O)
  - AI client management
  - Progress tracking
  - Error handling
- CLI-style invocation through main dispatcher
- Each tool is self-contained and testable

## Key Differences from Travel-Data Project

- **Travel-data**: Single-purpose (extract destinations)
- **This toolkit**: Prompt-driven and general-purpose
- **Travel-data**: Fixed transformation logic
- **This toolkit**: User defines transformation dynamically
- **Travel-data**: Hardcoded processing
- **This toolkit**: System adapts to any column structure and transformation need

## Use Cases

### Human Resources
"Add a column that extracts years of experience from job descriptions"

### Finance
"Categorize expenses into budget categories based on description"

### Communications
"Translate this column to French and add a sentiment score"

### Procurement
"Extract vendor names from descriptions and classify risk level"

### Project Management
"Summarize long status updates into brief 1-line summaries"

### Data Quality
"Identify and flag potential data quality issues in addresses"

## Benefits

This approach democratizes data processing with AI:
- No coding required
- Natural language descriptions
- Instant feedback loop
- Scalable processing
- Adaptable to any data transformation need
- Reduces dependency on technical teams
- Accelerates data workflow automation