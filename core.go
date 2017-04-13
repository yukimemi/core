package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// Cmd is command infomation.
type Cmd struct {
	Cmd     *exec.Cmd
	CmdLine string

	Stdin  bytes.Buffer
	Stdout bytes.Buffer
	Stderr bytes.Buffer

	ExitError error
	ExitCode  int

	StdinPipe  io.WriteCloser
	StdoutPipe io.ReadCloser
	StderrPipe io.ReadCloser

	StdinEnc  *encoding.Decoder
	StdoutEnc *encoding.Decoder
	StderrEnc *encoding.Decoder

	StdoutPrint bool
	StderrPrint bool

	wg sync.WaitGroup
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

// ScanPrintStdout is scan and print to stdout.
func ScanPrintStdout(scanner *bufio.Scanner, print bool) {
	for scanner.Scan() {
		if print {
			fmt.Fprintf(os.Stdout, "%s\n", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// ScanPrintStderr is scan and print to stderr.
func ScanPrintStderr(scanner *bufio.Scanner, print bool) {
	for scanner.Scan() {
		if print {
			fmt.Fprintf(os.Stderr, "%s\n", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// ReadToBuf is scan and store buffer.
func ReadToBuf(scanner *bufio.Scanner) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	for scanner.Scan() {
		buf.WriteString(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return buf, nil
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

// FailOnError is fail if err occured.
func FailOnError(err error) {
	if err != nil {
		log.Fatalf("%+v\n", err)
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
func (c *Cmd) CmdStart() error {

	var err error

	c.StdoutPipe, err = c.Cmd.StdoutPipe()
	if err != nil {
		return err
	}

	c.StderrPipe, err = c.Cmd.StderrPipe()
	if err != nil {
		return err
	}

	c.StdinPipe, err = c.Cmd.StdinPipe()
	if err != nil {
		return err
	}

	if c.StdoutEnc == nil {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			ScanPrintStdout(bufio.NewScanner(io.TeeReader(c.StdoutPipe, &c.Stdout)), c.StdoutPrint)
		}()
	} else {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			ScanPrintStdout(bufio.NewScanner(io.TeeReader(transform.NewReader(c.StdoutPipe, c.StdoutEnc), &c.Stdout)), c.StdoutPrint)
		}()
	}

	if c.StderrEnc == nil {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			ScanPrintStderr(bufio.NewScanner(io.TeeReader(c.StderrPipe, &c.Stderr)), c.StderrPrint)
		}()
	} else {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			ScanPrintStderr(bufio.NewScanner(io.TeeReader(transform.NewReader(c.StderrPipe, c.StderrEnc), &c.Stderr)), c.StderrPrint)
		}()
	}

	err = c.Cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

// CmdWait wait command end.
func (c *Cmd) CmdWait() {
	c.wg.Wait()
	c.ExitError = c.Cmd.Wait()
	c.GetExitCode()
}

// CmdRun run command.
func (c *Cmd) CmdRun() error {

	var err error

	err = c.CmdStart()
	if err != nil {
		return err
	}

	c.CmdWait()
	return nil
}

// GetExitCode return command ExitCode.
func (c *Cmd) GetExitCode() {
	if c.ExitError != nil {
		if err, ok := c.ExitError.(*exec.ExitError); ok {
			if s, ok := err.Sys().(syscall.WaitStatus); ok {
				c.ExitCode = s.ExitStatus()
			}
		}
	}
}

// SurroundWord surround word.
func SurroundWord(w string, r rune) string {

	surRune := []rune(w)

	// Check already surrounded.
	if surRune[0] != r {
		surRune = append([]rune{r}, surRune...)
	}
	if surRune[len(surRune)-1] != r {
		surRune = append(surRune, r)
	}

	return string(surRune)
}
