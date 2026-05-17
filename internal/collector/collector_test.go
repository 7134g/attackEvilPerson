package collector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadKeywords(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "kw_city.txt"), []byte("北京\n上海\n"), 0644)
	os.WriteFile(filepath.Join(tmp, "kw_hospital.txt"), []byte("内科\n外科\n"), 0644)

	keywords, err := LoadKeywords(tmp)
	if err != nil {
		t.Fatalf("LoadKeywords() error: %v", err)
	}
	if len(keywords) != 4 {
		t.Fatalf("expected 4 keywords, got %d", len(keywords))
	}
	expected := map[string]bool{
		"北京内科": true, "北京外科": true,
		"上海内科": true, "上海外科": true,
	}
	for _, kw := range keywords {
		if !expected[kw] {
			t.Errorf("unexpected keyword: %q", kw)
		}
	}
}

func TestAdPattern(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{`http://ada.baidu.com/site/jingshan-yanglao.com/xyl?imid=abc123`, true},
		{`https://ada.baidu.com/site/example.com/xyl?imid=xyz-789`, true},
		{`http://www.baidu.com/s/`, false},
		{`https://ada.baidu.com/other/`, false},
	}
	for _, c := range cases {
		got := adPattern.MatchString(c.input)
		if got != c.want {
			t.Errorf("MatchString(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestDedupe(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	got := dedupe(input)
	if len(got) != 3 {
		t.Errorf("dedupe len = %d, want 3", len(got))
	}
	seen := make(map[string]bool)
	for _, v := range got {
		if seen[v] {
			t.Errorf("duplicate found: %q", v)
		}
		seen[v] = true
	}
}

func TestUserAgentsNotEmpty(t *testing.T) {
	if len(userAgents) < 10 {
		t.Errorf("only %d user agents", len(userAgents))
	}
	for i, ua := range userAgents {
		if ua == "" {
			t.Errorf("user agent %d is empty", i)
		}
	}
}

func TestReadLines(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	os.WriteFile(path, []byte("line1\nline2\n\n  \nline3\n"), 0644)

	lines, err := readLines(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

func TestRun_DedupOutput(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "api.txt")

	// No keywords = no work, but should write empty file
	keywords := []string{}
	if err := Run(keywords, outputPath, Config{}); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(outputPath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// Should be empty or single empty line
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	if len(lines) != 0 {
		t.Errorf("expected empty output, got %d lines", len(lines))
	}
}
