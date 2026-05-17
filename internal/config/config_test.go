package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	data := `
baidu_url: 'https://www.baidu.com/'
tel_number: '13800000000'
tel_name: '测试'
proxy: 'http://127.0.0.1:7890'
browser_path: '/path/to/chrome'
titles:
  - "医生"
  - "护士"
relatives:
  - "我的父亲"
situations:
  - "生病了"
contact_methods:
  - "请联系 {number}"
greetings:
  - "您好"
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.BaiduURL != "https://www.baidu.com/" {
		t.Errorf("baidu_url = %q", cfg.BaiduURL)
	}
	if cfg.TelNumber != "13800000000" {
		t.Errorf("tel_number = %q", cfg.TelNumber)
	}
	if cfg.TelName != "测试" {
		t.Errorf("tel_name = %q", cfg.TelName)
	}
	if len(cfg.Titles) != 2 {
		t.Errorf("titles len = %d, want 2", len(cfg.Titles))
	}
	if len(cfg.ContactMethods) != 1 {
		t.Errorf("contact_methods len = %d, want 1", len(cfg.ContactMethods))
	}
}
