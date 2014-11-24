package gocore

import (
	"golang.org/x/text/encoding/japanese"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var tmp *os.File
var tmpfile string

func init() {
	// Make test csv.
	pwd, e := os.Getwd()
	FailOnError(e)
	f, e := ioutil.TempFile(pwd, "gocore")
	defer f.Close()

	f.WriteString("test01,test02,test03\n")
	f.WriteString("test11,test12\n")
	f.WriteString("test11,test12,test13,       test44\n")
	f.WriteString("# test21,test22,test23\n")
	f.Write([]byte{byte(0x93), byte(0xfa)})
	f.WriteString(",")
	f.Write([]byte{byte(0x93), byte(0xfa)})
	f.WriteString(",")
	f.Write([]byte{byte(0x93), byte(0xfa)})

	tmp = f
	tmpfile = f.Name()
}

func TestImportCsv(t *testing.T) {
	log.Printf(tmpfile)
	records, e := ImportCsv(tmpfile, japanese.ShiftJIS.NewDecoder())
	for r1 := range records {
		log.Print(records[r1])
	}

	if e != nil {
		t.Errorf(e.Error())
	}
}
