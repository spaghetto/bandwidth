package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type TestResult struct {
	Ping       Ping      `json:"ping"`
	Download   Bandwidth `json:"download"`
	Upload     Bandwidth `json:"upload"`
	PacketLoss float64   `json:"packetLoss"`

	ISP      string `json:"isp"`
	ErrorMsg string `json:"error"`
	Iface    Iface  `json:"interface"`
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

type Iface struct {
	Name       string
	ExternalIP string
}

func (r TestResult) Error() (bool, error) {
	if r.ErrorMsg == "" {
		return true, nil
	}

	return false, errors.New(r.ErrorMsg)
}

func Test(ctx context.Context, iface string) (*TestResult, error) {
	var result TestResult

	cmd := exec.CommandContext(ctx, "speedtest", // requires Ookla Speedtest, not speedtest-cli
		"-f", "json",
		"--accept-license", "--accept-gdpr", "--interface="+iface,
	)

	var sout, serr bytes.Buffer
	cmd.Stdout = &sout
	cmd.Stderr = &serr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = context.DeadlineExceeded
		}

		return nil, testErr{err: err, stderr: serr.String(), stdout: sout.String()}
	}

	if err := json.Unmarshal(sout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("Parsing speedtest output: '%w'", err)
	}

	if ok, err := result.Error(); !ok {
		return nil, testErr{err: err, stderr: serr.String(), stdout: sout.String()}
	}

	return &result, nil
}

type testErr struct {
	err    error
	stderr string
	stdout string
}

func (e testErr) Error() string {
	s := "Speedtest failed:\n"
	s += "  error  = " + indentStr(11, e.err.Error()) + "\n"

	e.stderr = strings.TrimSpace(e.stderr)
	if len(e.stderr) == 0 {
		e.stderr = "(empty)"
	}
	s += "  stderr = " + indentStr(11, e.stderr)

	e.stdout = strings.TrimSpace(e.stdout)
	if len(e.stdout) == 0 {
		e.stdout = "(empty)"
	}
	s += "  stdout = " + indentStr(11, e.stderr)

	return s
}

func indentStr(n int, s string) string {
	lines := strings.Split(s, "\n")

	for i := range lines {
		if i == 0 {
			continue
		}
		lines[i] = strings.Repeat(" ", n) + lines[i]
	}

	return strings.Join(lines, "\n")
}
