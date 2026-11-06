package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnthropicStructuredResponseAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing authentication")
		}
		var payload map[string]any
		if json.NewDecoder(r.Body).Decode(&payload) != nil {
			t.Error("invalid request")
		}
		if payload["model"] != "claude-sonnet-5-5" {
			t.Error("wrong model")
		}
		if _, ok := payload["temperature"]; ok {
			t.Error("unsupported sampling parameter")
		}
		if payload["output_config"].(map[string]any)["format"].(map[string]any)["type"] != "json_schema" {
			t.Error("missing structured format")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stop_reason":"end_turn","content":[{"type":"thinking","thinking":"private"},{"type":"text","text":"{\"direction\":\"LOWER_IS_BETTER\",\"response_kind\":\"NUMERIC\",\"explanation\":\"Less poverty is the stated objective.\"}"}],"usage":{"input_tokens":100,"output_tokens":20}}`))
	}))
	defer server.Close()
	c := newAnthropicClient("test-key", "claude-sonnet-5-5", "cpd-direction")
	c.endpoint = server.URL
	r, err := c.process(context.Background(), map[string]string{"definition": "Poverty rate"}, parseColumnSpecs("direction,response_kind,explanation"), cpdDirectionPrompt)
	if err != nil {
		t.Fatal(err)
	}
	if r.Results["direction"] != "LOWER_IS_BETTER" || r.InputTokens != 100 || r.OutputTokens != 20 || r.Tokens != 120 {
		t.Fatalf("incorrect result/usage: %+v", r)
	}
}

func TestRejectIncompleteAnswersAndDoNotExposeErrorBodies(t *testing.T) {
	for _, test := range []struct {
		name, body string
		status     int
	}{
		{"truncated", `{"stop_reason":"max_tokens","content":[{"type":"text","text":"{}"}]}`, 200},
		{"invalid direction", `{"stop_reason":"end_turn","content":[{"type":"text","text":"{\"direction\":\"GUESS\",\"response_kind\":\"NUMERIC\",\"explanation\":\"test\"}"}]}`, 200},
		{"categorical ratio", `{"stop_reason":"end_turn","content":[{"type":"text","text":"{\"direction\":\"HIGHER_IS_BETTER\",\"response_kind\":\"YES_NO\",\"explanation\":\"test\"}"}]}`, 200},
		{"authentication", `sensitive upstream error body`, 401},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(test.status); w.Write([]byte(test.body)) }))
			defer s.Close()
			c := newAnthropicClient("secret", "claude-sonnet-5-5", "cpd-direction")
			c.endpoint = s.URL
			_, err := c.process(context.Background(), nil, parseColumnSpecs("direction,response_kind,explanation"), cpdDirectionPrompt)
			if err == nil {
				t.Fatal("accepted incomplete/invalid answer")
			}
			if strings.Contains(err.Error(), "sensitive") || strings.Contains(err.Error(), "secret") {
				t.Fatal("leaked response or credentials")
			}
		})
	}
}

func TestRaggedExcelRowsKeepGeneratedColumnsAligned(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"stop_reason":"end_turn","content":[{"type":"text","text":"{\"label\":\"ok\"}"}]}`))
	}))
	defer s.Close()
	c := newAnthropicClient("test", "claude-sonnet-5-5", "")
	c.endpoint = s.URL
	rows, stats := processFullDataset(context.Background(), c, []string{"id", "optional"}, [][]string{{"1"}, {"2", "context"}}, parseColumnSpecs("label"), "Classify", 2, 100, t.TempDir()+"/output.csv")
	if stats.CompletedRows != 2 || len(rows[0]) != 3 || rows[0][1] != "" || rows[0][2] != "ok" {
		t.Fatalf("misaligned results: %v", rows)
	}
}
