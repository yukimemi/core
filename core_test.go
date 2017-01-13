package core

import (
	"os"
	"testing"
)

// TestIsMatch is Test IsMatchStrs func.
func TestIsMatch(t *testing.T) {
	var (
		err     error
		ma      bool
		matches []string
	)

	matches = []string{
		"hoge",
		"fuga",
		"hage",
	}

	ma, err = IsMatchStrs("hoge", matches)
	if err != nil {
		t.FailNow()
	}
	if !ma {
		t.FailNow()
	}
	ma, err = IsMatchStrs("fuga", matches)
	if err != nil {
		t.FailNow()
	}
	if !ma {
		t.FailNow()
	}
	ma, err = IsMatchStrs("hage", matches)
	if err != nil {
		t.FailNow()
	}
	if !ma {
		t.FailNow()
	}
	ma, err = IsMatchStrs("moge", matches)
	if err != nil {
		t.FailNow()
	}
	if ma {
		t.FailNow()
	}

	matches = []string{}
	ma, err = IsMatchStrs("foo|bar", matches)
	if err != nil {
		t.FailNow()
	}
	if !ma {
		t.FailNow()
	}

	matches = make([]string, 0)
	ma, err = IsMatchStrs("hogehoge", matches)
	if err != nil {
		t.FailNow()
	}
	if !ma {
		t.FailNow()
	}

}

// TestMain is entry point.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
