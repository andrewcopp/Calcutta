package lab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

const defaultGenerateEntriesTimeout = 5 * time.Minute

// pythonGenerateEntriesResult matches the JSON output from generate_lab_entries.py --json-output
type pythonGenerateEntriesResult struct {
	OK             bool     `json:"ok"`
	EntriesCreated int      `json:"entries_created"`
	Errors         []string `json:"errors"`
}

// RunGenerateEntriesScript executes the Python script to generate lab entries.
func RunGenerateEntriesScript(ctx context.Context, modelID string, req models.LabGenerateEntriesRequest) (*models.LabGenerateEntriesResponse, error) {
	pythonBin := strings.TrimSpace(os.Getenv("PYTHON_BIN"))
	if pythonBin == "" {
		pythonBin = "python3"
	}

	scriptPath := strings.TrimSpace(os.Getenv("PYTHON_GENERATE_LAB_ENTRIES"))
	candidates := make([]string, 0, 2)
	if scriptPath != "" {
		candidates = append(candidates, scriptPath)
	}
	candidates = append(candidates,
		"data-science/scripts/generate_lab_entries.py",
		"../data-science/scripts/generate_lab_entries.py",
	)

	resolvedScript := ""
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			resolvedScript = abs
			break
		}
	}
	if resolvedScript == "" {
		return nil, errors.New("generate_lab_entries.py not found; set PYTHON_GENERATE_LAB_ENTRIES")
	}

	// Build arguments
	args := []string{
		resolvedScript,
		"--model-id", modelID,
		"--json-output",
	}

	if len(req.Years) > 0 {
		yearStrs := make([]string, len(req.Years))
		for i, y := range req.Years {
			yearStrs[i] = strconv.Itoa(y)
		}
		args = append(args, "--years", strings.Join(yearStrs, ","))
	}

	if req.BudgetPoints > 0 {
		args = append(args, "--budget", strconv.Itoa(req.BudgetPoints))
	}

	if req.ExcludedEntry != "" {
		args = append(args, "--excluded-entry", req.ExcludedEntry)
	}

	// Apply timeout
	timeout := defaultGenerateEntriesTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pythonBin, args...)
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := strings.TrimSpace(stdout.String())

	// Parse JSON output
	var parsed pythonGenerateEntriesResult
	if outStr != "" {
		_ = json.Unmarshal([]byte(outStr), &parsed)
	}

	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(parsed.Errors) > 0 {
			msg = strings.Join(parsed.Errors, "; ")
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}

	if !parsed.OK {
		msg := "python script returned ok=false"
		if len(parsed.Errors) > 0 {
			msg = strings.Join(parsed.Errors, "; ")
		}
		return nil, errors.New(msg)
	}

	return &models.LabGenerateEntriesResponse{
		EntriesCreated: parsed.EntriesCreated,
		Errors:         parsed.Errors,
	}, nil
}
