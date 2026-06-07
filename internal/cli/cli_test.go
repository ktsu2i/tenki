package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsStubSummary(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run([]string{"tokyo"}, &stdout, &stderr, "test")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	out := stdout.String()
	for _, want := range []string{"tenki: tokyo", "mode: summary", "weather fetching is not implemented yet"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout = %q, want it to contain %q", out, want)
		}
	}
}

func TestRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run([]string{"tokyo", "--hourly", "--hours", "24", "--json"}, &stdout, &stderr, "test")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	want := "{\"location\":\"tokyo\",\"mode\":\"hourly\",\"days\":3,\"hours\":24,\"status\":\"stub\"}\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunRejectsConflictingModes(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run([]string{"tokyo", "--daily", "--hourly"}, &stdout, &stderr, "test")
	if err == nil {
		t.Fatal("Run returned nil error, want conflict error")
	}
	if !strings.Contains(err.Error(), "--daily and --hourly cannot be used together") {
		t.Fatalf("error = %q, want mode conflict", err.Error())
	}
}

func TestRunRejectsInvalidDays(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run([]string{"tokyo", "--days", "0"}, &stdout, &stderr, "test")
	if err == nil {
		t.Fatal("Run returned nil error, want invalid days error")
	}
	if !strings.Contains(err.Error(), "--days must be between 1 and 7") {
		t.Fatalf("error = %q, want days range error", err.Error())
	}
}

func TestRunRequiresLocation(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run(nil, &stdout, &stderr, "test")
	if err == nil {
		t.Fatal("Run returned nil error, want location error")
	}
	if !strings.Contains(err.Error(), "<location>") {
		t.Fatalf("error = %q, want location error", err.Error())
	}
}
