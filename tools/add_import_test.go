package tools

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestAddImport(t *testing.T) {
	for _, tt := range []struct {
		Name     string
		Filename string
	}{
		{"Empty import decl", "test_empty_add_import.go"},
		{"One import decl", "test_one_add_import.go"},
		{"Many import decl", "test_many_add_import.go"},
		{"Many with error import decl", "test_many_with_error_add_import.go"},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			testAddImport(t, tt.Filename)
		})
	}
}

func testAddImport(t *testing.T, filename string) {
	fullFilename := "./testdata/" + filename
	goldenFullFilename := "./testdata/" + filename + ".golden"

	prevBs, err := ioutil.ReadFile(fullFilename)
	if err != nil {
		t.Fatalf("Error on read prev version of file: %v", err)
	}

	defer func() {
		err = ioutil.WriteFile(fullFilename, prevBs, 0666)
		if err != nil {
			t.Errorf("Error on revert result file: %v", err)
		}
	}()

	err = addImportToFile(fullFilename, "bytes", prevBs)
	if err != nil {
		t.Errorf("Error on add import: %v", err)
	}

	resultBs, err := ioutil.ReadFile(fullFilename)
	if err != nil {
		t.Errorf("Error on read result file (%v): %v", fullFilename, err)
	}

	goldenBs, err := ioutil.ReadFile(goldenFullFilename)
	if err != nil {
		t.Errorf("Error on read golden file (%v): %v", goldenFullFilename, err)
	}

	if !bytes.Equal(resultBs, goldenBs) {
		t.Errorf("Result of %v", fullFilename)
		t.Errorf("Result: %v", string(resultBs))
		t.Errorf("Expect: %v", string(goldenBs))
		t.Errorf("Result not correct")
	}
}

func addImportToFile(filename string, importName string, prevBs []byte) error { // nolint: gocyclo
	res, err := AddImport(filename, importName)
	if err != nil {
		return errors.Wrap(err, "error on add import")
	}

	out, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "error on create temp file")
	}

	n, err := out.Write(prevBs[:res.Lpos])
	if err != nil {
		return errors.Wrap(err, "error on copy first part")
	}

	if n != int(res.Lpos) {
		return errors.New("wrong copied bytes")
	}

	_, err = out.WriteString(res.Text)
	if err != nil {
		return errors.Wrap(err, "error on write text of new import block")
	}

	_, err = out.Write(prevBs[res.Rpos:])
	if err != nil {
		return errors.Wrap(err, "error on copy file")
	}

	err = out.Close()
	if err != nil {
		return errors.Wrap(err, "error on close to read")
	}

	return nil
}
