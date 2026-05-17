package message

import (
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	tmpl := Templates{
		Titles:         []string{"医生"},
		Relatives:      []string{"我的父亲"},
		Situations:     []string{"生病了"},
		ContactMethods: []string{"请联系 {number}"},
		Greetings:      []string{"您好"},
	}

	msg := Build(tmpl, "13800000000", "张伟")
	if !strings.HasPrefix(msg, "您好") {
		t.Errorf("message should start with greeting: %s", msg)
	}
	if !strings.Contains(msg, "13800000000") {
		t.Errorf("message should contain phone number: %s", msg)
	}
	if !strings.Contains(msg, "张伟") {
		t.Errorf("message should contain tel_name: %s", msg)
	}
	if strings.Contains(msg, "{number}") {
		t.Errorf("message should not contain unreplaced placeholder: %s", msg)
	}
}

func TestBuildRandomness(t *testing.T) {
	tmpl := Templates{
		Titles:         []string{"医生", "护士", "主任"},
		Relatives:      []string{"我的父亲", "我的母亲"},
		Situations:     []string{"生病了", "需要帮助"},
		ContactMethods: []string{"请联系 {number}"},
		Greetings:      []string{"您好", "你好"},
	}

	// Run many times to ensure no panic and reasonable output
	for i := 0; i < 100; i++ {
		msg := Build(tmpl, "13800000000", "张伟")
		if msg == "" {
			t.Fatal("got empty message")
		}
	}
}
