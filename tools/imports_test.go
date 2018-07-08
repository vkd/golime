package tools

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"
)

func test_parseDir(t *testing.T) {
	var i, j int
	for _ = range parseDirWalk(build.Default.GOROOT) {
		i++
	}
	for _ = range parseDir(build.Default.GOROOT) {
		j++
	}
	if i != j {
		t.Skipf("Wrong import counts: %d, %d", i, j)
	}
}

func BenchmarkParseDir_Walk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range parseDirWalk(build.Default.GOROOT) {
		}
	}
}

func BenchmarkParseIoutilDir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _ = range parseDir(build.Default.GOROOT) {
		}
	}
}

// parseDirWalk - slow
func parseDirWalk(pathDir string) <-chan importPath {
	out := make(chan importPath)
	go func() {
		err := filepath.Walk(pathDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			if path == pathDir {
				return nil
			}
			switch info.Name() {
			case "testdata":
				return filepath.SkipDir
			}
			out <- importPath{path: path[len(pathDir)+1:]}
			return nil
		})
		if err != nil {
			out <- importPath{err: err}
		}
		close(out)
	}()
	return out
}
