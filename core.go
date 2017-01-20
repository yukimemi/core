package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Cmd is command infomation.
type Cmd struct {
	File string
	Dir  string
	Name string
	Cwd  string
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
func GetCmdInfo() (Cmd, error) {

	var (
		ci  Cmd
		err error
	)

	// Get cmd info.
	cmdFile, err := GetCmdPath(os.Args[0])
	if err != nil {
		return ci, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ci, err
	}
	cmdDir := filepath.Dir(cmdFile)
	return Cmd{
		File: cmdFile,
		Dir:  cmdDir,
		Name: BaseName(cmdFile),
		Cwd:  cwd,
	}, nil
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
