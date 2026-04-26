package version

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	// 保存原始值
	origVersion := Version
	origCommit := GitCommit
	origDate := BuildDate
	defer func() {
		Version = origVersion
		GitCommit = origCommit
		BuildDate = origDate
	}()

	Version = "v1.0.0"
	GitCommit = "abc1234"
	BuildDate = "2024-01-01T00:00:00Z"

	var buf bytes.Buffer
	Print(&buf)

	output := buf.String()

	expectations := []struct {
		label string
		want  string
	}{
		{"版本号", "v1.0.0"},
		{"Git 提交", "abc1234"},
		{"构建日期", "2024-01-01T00:00:00Z"},
		{"Go 版本", runtime.Version()},
	}

	for _, e := range expectations {
		if !strings.Contains(output, e.want) {
			t.Errorf("Print 输出中缺少 %s: %q，完整输出:\n%s", e.label, e.want, output)
		}
	}
}

func TestPrintDefaultValues(t *testing.T) {
	// 保存原始值
	origVersion := Version
	origCommit := GitCommit
	origDate := BuildDate
	defer func() {
		Version = origVersion
		GitCommit = origCommit
		BuildDate = origDate
	}()

	Version = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	var buf bytes.Buffer
	Print(&buf)

	output := buf.String()
	if !strings.Contains(output, "dev") {
		t.Error("默认版本应包含 'dev'")
	}
	if !strings.Contains(output, "unknown") {
		t.Error("默认值应包含 'unknown'")
	}
}
