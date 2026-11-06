package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

func saveRunManifest(output, input, prompt string, c *anthropicClient, stats *ProcessingStats) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	inputHash := sha256.Sum256(data)
	promptHash := sha256.Sum256([]byte(prompt))
	record := map[string]any{
		"provider": "anthropic", "model": c.model, "preset": c.preset,
		"input_sha256": hex.EncodeToString(inputHash[:]), "prompt_sha256": hex.EncodeToString(promptHash[:]),
		"prompt": prompt, "created_at": time.Now().UTC().Format(time.RFC3339),
		"selected_rows": stats.TotalRows, "completed_rows": stats.CompletedRows, "failed_rows": stats.FailedRows,
		"input_tokens": stats.InputTokens, "output_tokens": stats.OutputTokens,
		"review_status": "unreviewed",
	}
	data, err = json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(output+".run.json", append(data, '\n'), 0600)
}
