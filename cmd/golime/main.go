package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/cweill/gotests"
	"github.com/pkg/errors"
	"github.com/vkd/golime/tools"
)

var (
	version = "v0.0.0.2"
)

type Cmd func(data []byte) (out interface{}, err error)

func (c Cmd) Run(data []byte) (out interface{}, err error) {
	if c == nil {
		return nil, fmt.Errorf("not implemented")
	}
	return c(data)
}

func Run(i int) bool {
	panic("Not implemented")
}

type Result map[string]interface{}

var commands = map[string]Cmd{
	"version": func(data []byte) (out interface{}, err error) {
		return Result{"version": version}, nil
	},
	"imports": func(data []byte) (out interface{}, err error) {
		imports, err := tools.GetAllImportPaths()
		if err != nil {
			return nil, err
		}
		return Result{"imports": imports}, nil
	},
	"add_import": func(data []byte) (out interface{}, err error) {
		type st struct {
			Import string `json:"import"`
			File   string `json:"file"`
		}
		var s st
		err = json.Unmarshal(data, &s)
		if err != nil {
			return nil, errors.Wrap(err, "error on umarshal data")
		}
		res, err := tools.AddImport(s.File, s.Import)
		if err != nil {
			return nil, errors.Wrap(err, "error on add import")
		}
		return Result{"status": "ok", "result": res}, nil
	},
	"gotest": func(data []byte) (out interface{}, err error) {
		type st struct {
			File         string `json:"file"`
			FunctionName string `json:"function"`
		}
		var s st
		err = json.Unmarshal(data, &s)
		if err != nil {
			return nil, errors.Wrap(err, "error on unmarshal data")
		}
		rgx, err := regexp.Compile("(?i)^(" + s.FunctionName + ")$")
		if err != nil {
			return nil, errors.Wrap(err, "error on compile regexp")
		}
		ts, err := gotests.GenerateTests(s.File, &gotests.Options{
			Only: rgx,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error on get tests")
		}
		var paths []string
		for _, t := range ts {
			err = ioutil.WriteFile(t.Path, t.Output, 0644)
			if err != nil {
				return nil, errors.Wrapf(err, "error on write test result (%v)", t.Path)
			}
			paths = append(paths, t.Path)
		}
		return Result{"test_files": paths}, nil
	},
	// "goiface": func(data []byte) (out interface{}, err error) {
	// 	type st struct {
	// 		Receiver string `json:"receiver"`
	// 		Iface    string `json:"iface"`
	// 		Filepath string `json:"filepath"`
	// 	}
	// 	var s st
	// 	err = json.Unmarshal(data, &s)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "error on unmarshal struct")
	// 	}

	// 	file, err := os.OpenFile(s.Filepath, os.O_WRONLY, 0600)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "error on open file")
	// 	}
	// 	defer file.Close()

	// 	_, err = file.Seek(0, os.SEEK_END)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "error on seek")
	// 	}

	// 	err = goiface.Impl(s.Receiver, s.Iface, file)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "error on impl")
	// 	}
	// 	panic("Not implemented")
	// },
}

type CmdArgs struct {
	Cmd  string          `json:"cmd"`
	Data json.RawMessage `json:"data"`
}

var (
	versionFlag = flag.Bool("v", false, "Version of golime")
)

func main() {
	flag.Parse()
	if *versionFlag {
		log.Printf("Version: %v", version)
		return
	}

	http.HandleFunc("/cmd", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var cmd CmdArgs
		err := json.NewDecoder(req.Body).Decode(&cmd)
		if err != nil {
			writeError(w, err)
			return
		}
		out, err := commands[cmd.Cmd].Run(cmd.Data)
		if err != nil {
			writeError(w, err)
			return
		}

		json.NewEncoder(w).Encode(out)
	}))
	http.HandleFunc("/stop", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		os.Exit(0)
	}))
	log.Fatal(http.ListenAndServe(":8601", nil))

	// goroot := build.Default.GOROOT
	// var count int
	// for p := range parseDir(path.Join(goroot, "src"), skipVendor) {
	// 	if p.err != nil {
	// 		log.Fatalf("Error on parse dir (%s): %v", goroot, p.err)
	// 	}
	// 	fmt.Println(p.path)
	// 	count++
	// }
	// log.Printf("count: %d", count)
	// for p := range parseDir(path.Join(build.Default.GOPATH, "src")) {
	// 	if p.err != nil {
	// 		log.Fatalf("Error on parse dir (%s): %v", goroot, p.err)
	// 	}
	// 	fmt.Println(p.isVendor, p.path)
	// 	count++
	// }
	// log.Printf("count: %d", count)
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
}
