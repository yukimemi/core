package gocore

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"

	"github.com/umisama/golog"
	"golang.org/x/text/transform"
)

var Logger log.Logger

func FailOnError(e error) { // {{{
	if e != nil {
		Logger.Critical("Error:", e)
	}
} // }}}

func GetBaseName(path string) string { // {{{
	_, base := filepath.Split(path)
	ext := filepath.Ext(base)
	basename := strings.TrimSuffix(base, ext)
	return basename
} // }}}

func ImportCsv(path string, enc transform.Transformer) ([][]string, error) { // {{{
	file, e := os.Open(path)
	FailOnError(e)
	defer file.Close()

	var r *csv.Reader
	if enc != nil {
		tr := transform.NewReader(file, enc)
		r = csv.NewReader(tr)
	} else {
		r = csv.NewReader(file)
	}
	r.Comment = '#'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	records, e := r.ReadAll()
	FailOnError(e)
	return records, e
} // }}}
