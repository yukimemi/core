package gocore

import (
	"path/filepath"
	"strings"

	"github.com/umisama/golog"
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
