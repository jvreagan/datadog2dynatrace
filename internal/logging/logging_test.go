package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	t.Cleanup(func() {
		SetLevel(LevelWarn)
		SetOutput(nil) // reset isn't critical; tests are isolated
	})

	tests := []struct {
		name      string
		level     Level
		logFunc   func(string, ...interface{})
		wantEmpty bool
	}{
		{"error at warn level", LevelWarn, Errorf, false},
		{"warn at warn level", LevelWarn, Warn, false},
		{"info at warn level", LevelWarn, Info, true},
		{"debug at warn level", LevelWarn, Debug, true},
		{"info at info level", LevelInfo, Info, false},
		{"debug at info level", LevelInfo, Debug, true},
		{"debug at debug level", LevelDebug, Debug, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			SetLevel(tt.level)
			tt.logFunc("test message")
			got := buf.String()
			if tt.wantEmpty && got != "" {
				t.Errorf("expected no output, got %q", got)
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("expected output, got nothing")
			}
		})
	}
}

func TestOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	t.Cleanup(func() {
		SetLevel(LevelWarn)
	})

	Errorf("err %d", 1)
	Warn("warn %s", "msg")
	Info("info")
	Debug("debug val=%v", true)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d: %v", len(lines), lines)
	}

	expected := []string{
		"[ERROR] err 1",
		"[WARN] warn msg",
		"[INFO] info",
		"[DEBUG] debug val=true",
	}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want)
		}
	}
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	t.Cleanup(func() {
		SetLevel(LevelWarn)
	})

	w := Writer()
	n, err := w.Write([]byte("hello from writer\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 18 {
		t.Errorf("expected n=18, got %d", n)
	}

	got := buf.String()
	if !strings.Contains(got, "[DEBUG] hello from writer") {
		t.Errorf("expected debug log from writer, got %q", got)
	}
}

func TestWriterSuppressedAtWarnLevel(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelWarn)

	w := Writer()
	w.Write([]byte("should not appear\n"))

	if buf.String() != "" {
		t.Errorf("expected no output at warn level, got %q", buf.String())
	}
}
