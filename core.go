package core

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// Cmd is command infomation.
type Cmd struct {
	File string
	Dir  string
	Name string
	Cwd  string

	Args    []string
	CmdLine string

	Info   os.FileInfo
	ExeCmd *exec.Cmd
	StdIn  io.WriteCloser
	StdOut io.ReadCloser
	StdErr io.ReadCloser

	StdInEnc  *encoding.Decoder
	StdOutEnc *encoding.Decoder
	StdErrEnc *encoding.Decoder
}

// IsMatchStrs is whether str match or not.
func IsMatchStrs(str string, regStrs []string) (bool, error) {

	var err error

	if len(regStrs) == 0 {
		return true, nil
	}
	re, err := CompileStrs(regStrs)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), nil
}

// CompileStrs is regexp strings compile to *regexp.Regexp.
func CompileStrs(regStrs []string) (*regexp.Regexp, error) {
	if len(regStrs) == 0 {
		return nil, fmt.Errorf("regStrs must be greater than or equal to 1")
	}
	regStr := "(" + strings.Join(regStrs, "|") + ")"
	re, err := regexp.Compile(regStr)
	if err != nil {
		return nil, err
	}
	return re, nil
}

// ScanLoop is scan and print.
func ScanLoop(scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// BaseName is get file name without extension.
func BaseName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	re := regexp.MustCompile(ext + "$")
	return re.ReplaceAllString(base, "")
}

// GetCmdPath returns cmd abs path.
func GetCmdPath(cmd string) (string, error) {
	if filepath.IsAbs(cmd) {
		return cmd, nil
	}
	return exec.LookPath(cmd)
}

// GetCmdInfo return struct of Cmd.
func GetCmdInfo(path string) (Cmd, error) {

	var (
		err error
		ci  Cmd
	)

	ci.File, err = GetCmdPath(path)
	if err != nil {
		return ci, err
	}
	ci.Cwd, err = os.Getwd()
	if err != nil {
		return ci, err
	}
	ci.Dir = filepath.Dir(ci.File)
	ci.Name = BaseName(ci.File)

	ci.Info, err = os.Stat(ci.File)
	if err != nil {
		return ci, err
	}

	return ci, nil
}

// FailOnError is fail if err occured.
func FailOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// GetGlobArgs return glob files.
func GetGlobArgs(args []string) ([]string, error) {

	var a []string

	for _, v := range args {
		files, err := filepath.Glob(v)
		if err != nil {
			return nil, err
		}
		a = append(a, files...)
	}

	return a, nil
}

// CmdStart start cmdFile.
func CmdStart(cmd Cmd) (Cmd, error) {

	var err error

	cmd.ExeCmd = exec.Command(cmd.File, cmd.Args...)
	cmd.StdOut, err = cmd.ExeCmd.StdoutPipe()
	if err != nil {
		return cmd, err
	}
	cmd.StdErr, err = cmd.ExeCmd.StderrPipe()
	if err != nil {
		return cmd, err
	}
	cmd.StdIn, err = cmd.ExeCmd.StdinPipe()
	if err != nil {
		return cmd, err
	}

	if cmd.StdOutEnc == nil {
		go ScanLoop(bufio.NewScanner(cmd.StdOut))
	} else {
		go ScanLoop(bufio.NewScanner(transform.NewReader(cmd.StdOut, cmd.StdOutEnc)))
	}

	if cmd.StdErrEnc == nil {
		go ScanLoop(bufio.NewScanner(cmd.StdErr))
	} else {
		go ScanLoop(bufio.NewScanner(transform.NewReader(cmd.StdErr, cmd.StdErrEnc)))
	}

	/* fmt.Println("Exec:", cmd.File, "Args:", cmd.Args) */
	err = cmd.ExeCmd.Start()
	if err != nil {
		return cmd, err
	}

	return cmd, nil
}
