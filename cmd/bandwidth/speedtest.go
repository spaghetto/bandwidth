package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type TestResult struct {
	Ping     Ping      `json:"ping"`
	Download Bandwidth `json:"download"`
	Upload   Bandwidth `json:"upload"`

	ISP      string `json:"isp"`
	ErrorMsg string `json:"error"`
}

type Ping struct {
	Latency float64 `json:"latency"`
	Jitter  float64 `json:"jitter"`
}

type Bandwidth struct {
	Bandwidth float64 `josn:"bandwidth"`
	Bytes     float64 `josn:"bytes"`
	Elapsed   float64 `josn:"elapsed"`
}

func (r TestResult) Error() (bool, error) {
	if r.ErrorMsg == "" {
		return true, nil
	}

	return false, errors.New(r.ErrorMsg)
}

func Test(serverID string) (*TestResult, error) {
	var result TestResult

	cmd := exec.Command("speedtest", // requires Ookla Speedtest, not speedtest-cli
		"-f", "json",
		"--accept-license", "--accept-gdpr",
	)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Speedtest failed: '%w'", err)
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("Parsing speedtest output: '%w'", err)
	}

	if ok, err := result.Error(); !ok {
		return nil, err
	}

	return &result, nil
}
