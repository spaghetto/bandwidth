package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type TestResult struct {
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`
	Ping     float64 `json:"ping"`
}

func Test(serverID string) (*TestResult, error) {
	var result TestResult

	cmd := exec.Command("speedtest", "--json")

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Speedtest failed: '%w'", err)
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("Parsing speedtest output: '%w'", err)
	}

	return &result, nil
}
