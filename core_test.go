package core

import (
	"os"
	"os/exec"
	"runtime"
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

// TestBaseName is test BaseName fucn.
func TestBaseName(t *testing.T) {
	p := "/path/to/file.txt"
	e := "file"

	a := BaseName(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	p = "/path/to/file.txt.ext"
	e = "file.txt"

	a = BaseName(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	p = "/パス/トゥ/日本語パス.txt.ext"
	e = "日本語パス.txt"

	a = BaseName(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}
}

// TestGetCmdPath test GetCmdPath func.
func TestGetCmdPath(t *testing.T) {
	p := "go"
	e, err := exec.LookPath("go")
	if err != nil {
		t.Fail()
	}

	a, err := GetCmdPath(p)
	if err != nil {
		t.Fail()
	}
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	if runtime.GOOS == "windows" {
		p = "C:\\bin\\go"
		e = "C:\\bin\\go"
	} else {
		p = "/opt/local/bin/go"
		e = "/opt/local/bin/go"
	}

	a, err = GetCmdPath(p)
	if err != nil {
		t.Fail()
	}
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}
}

// TestMain is entry point.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
