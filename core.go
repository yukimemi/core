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

	"github.com/k0kubun/pp"
	"github.com/mattn/go-shellwords"
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

	StdinEnc  *encoding.Decoder
	StdoutEnc *encoding.Decoder
	StderrEnc *encoding.Decoder

	StdoutPrint bool
	StderrPrint bool

	stdinPipe  bool
	stdoutPipe bool
	stderrPipe bool
	Wg         sync.WaitGroup
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

// ScanWrite is scan and write to io.Writer.
func ScanWrite(scanner *bufio.Scanner, writer io.Writer, print bool) {
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Fprintf(writer, "%s\n", t)
		if print {
			fmt.Println(t)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
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

// NewCmd create Cmd struct pointer.
func NewCmd(cmdLine string) (*Cmd, error) {

	parseCmd, err := shellwords.Parse(cmdLine)
	if err != nil {
		return nil, err
	}
	return &Cmd{Cmd: exec.Command(parseCmd[0], parseCmd[1:]...)}, nil
}

// StdoutPipe return exec.StdoutPipe.
func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	c.stdoutPipe = true
	return c.Cmd.StdoutPipe()
}

// StderrPipe return exec.StderrPipe.
func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	c.stderrPipe = true
	return c.Cmd.StderrPipe()
}

// StdinPipe return exec.StdinPipe.
func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	c.stdinPipe = true
	return c.Cmd.StdinPipe()
}

// StdoutScanner make bufio.Scanner.
func (c *Cmd) StdoutScanner() (*bufio.Scanner, error) {
	c.stdoutPipe = true
	r, err := c.Cmd.StdoutPipe()
	if c.StdoutEnc != nil {
		if err != nil {
			return nil, err
		}
		return bufio.NewScanner(transform.NewReader(r, c.StdoutEnc)), nil
	}
	return bufio.NewScanner(r), nil
}

// StderrScanner make bufio.Scanner.
func (c *Cmd) StderrScanner() (*bufio.Scanner, error) {
	c.stderrPipe = true
	r, err := c.Cmd.StderrPipe()
	if c.StderrEnc != nil {
		if err != nil {
			return nil, err
		}
		return bufio.NewScanner(transform.NewReader(r, c.StderrEnc)), nil
	}
	return bufio.NewScanner(r), nil
}

// CmdStart start cmdFile.
func (c *Cmd) CmdStart() error {

	var err error

	// CmdLine check
	if c.CmdLine != "" && c.Cmd == nil {
		parseCmd, err := shellwords.Parse(c.CmdLine)
		if err != nil {
			return err
		}
		c.Cmd = exec.Command(parseCmd[0], parseCmd[1:]...)
	}
	if c.CmdLine == "" {
		c.CmdLine = strings.Join(c.Cmd.Args, " ")
	}

	var stdoutPipe, stderrPipe io.ReadCloser

	if !c.stdoutPipe {
		stdoutPipe, err = c.Cmd.StdoutPipe()
		if err != nil {
			return err
		}
	}
	if !c.stderrPipe {
		stderrPipe, err = c.Cmd.StderrPipe()
		if err != nil {
			return err
		}
	}

	// Start command.
	err = c.Cmd.Start()
	if err != nil {
		c.ExitError = err
		c.GetExitCode()
		return err
	}

	scanWrite := func(c *Cmd, r io.ReadCloser, w io.Writer, enc *encoding.Decoder, print bool) {
		if r == nil {
			return
		}
		c.Wg.Add(1)
		go func() {
			defer c.Wg.Done()
			if enc == nil {
				ScanWrite(bufio.NewScanner(r), w, print)
			} else {
				ScanWrite(bufio.NewScanner(transform.NewReader(r, enc)), w, print)
			}
		}()
	}

	scanWrite(c, stdoutPipe, &c.Stdout, c.StdoutEnc, c.StdoutPrint)
	scanWrite(c, stderrPipe, &c.Stderr, c.StderrEnc, c.StderrPrint)

	return nil
}

// CmdWait wait command end.
func (c *Cmd) CmdWait() {
	c.Wg.Wait()
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

// GetExitCode set command ExitCode.
func (c *Cmd) GetExitCode() {

	c.ExitCode = 1

	if c.ExitError != nil {
		if c.Cmd.ProcessState != nil {
			if status, ok := c.Cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
				c.ExitCode = status.ExitStatus()
			}
		}
	} else {
		c.ExitCode = 0
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

// PP is wrapper of pp Println.
func PP(a ...interface{}) (int, error) {
	return pp.Println(a...)
}

// PPf is wrapper of pp Printf.
func PPf(format string, a ...interface{}) (int, error) {
	return pp.Printf(format, a...)
}
