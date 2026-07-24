package benchmark

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"skilltrace/internal/adapters"
	"skilltrace/internal/adapters/claude"
	"skilltrace/internal/adapters/codex"
	"skilltrace/internal/adapters/huggingface"
	"skilltrace/internal/trace"
	"sort"
)

type fixture struct {
	Harness, Model, Tier, Trace string
	ExpectedEvents              int `json:"expected_events"`
}

func Run(ctx context.Context, root string) (Report, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Report{}, fmt.Errorf("read fixtures: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	r := adapters.NewRegistry()
	sanitizer := trace.NewHashSanitizer("skilltrace-v1")
	r.Register("codex", codex.New(sanitizer))
	r.Register("claude", claude.New(sanitizer))
	r.Register("huggingface", huggingface.New(sanitizer))
	report := Report{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		f, err := os.Open(filepath.Join(root, entry.Name()))
		if err != nil {
			return Report{}, err
		}
		var spec fixture
		err = json.NewDecoder(f).Decode(&spec)
		f.Close()
		if err != nil {
			return Report{}, fmt.Errorf("fixture %s: %w", entry.Name(), err)
		}
		traceFile, err := os.Open(filepath.Join(root, spec.Trace))
		if err != nil {
			return Report{}, err
		}
		result, err := r.Parse(ctx, spec.Harness, bufio.NewReader(traceFile))
		traceFile.Close()
		if err != nil {
			return Report{}, err
		}
		matched := len(result.Events)
		expected := spec.ExpectedEvents
		precision, recall := 1.0, 1.0
		if matched > expected && matched > 0 {
			precision = float64(expected) / float64(matched)
		}
		if matched < expected && expected > 0 {
			recall = float64(matched) / float64(expected)
		}
		report.Fixtures++
		report.Slices = append(report.Slices, Slice{Harness: spec.Harness, Model: spec.Model, Tier: spec.Tier, Expected: expected, Matched: matched, Precision: precision, Recall: recall})
	}
	return report, nil
}
