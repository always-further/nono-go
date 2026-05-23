package nono_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/always-further/nono-go"
)

func ExampleQueryContext() {
	dir, err := os.MkdirTemp("", "nono-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	caps := nono.New()
	defer caps.Close()
	if err := caps.AllowPath(dir, nono.AccessRead); err != nil {
		panic(err)
	}
	if err := caps.SetNetworkMode(nono.NetworkAllowAll); err != nil {
		panic(err)
	}

	qc, err := nono.NewQueryContext(caps)
	if err != nil {
		panic(err)
	}
	defer qc.Close()

	pathResult, err := qc.QueryPath(filepath.Join(dir, "file.txt"), nono.AccessRead)
	if err != nil {
		panic(err)
	}
	networkResult, err := qc.QueryNetwork()
	if err != nil {
		panic(err)
	}

	fmt.Println(pathResult.Status)
	fmt.Println(networkResult.Status)

	// Output:
	// allowed
	// allowed
}

func ExampleSandboxState() {
	dir, err := os.MkdirTemp("", "nono-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	caps := nono.New()
	defer caps.Close()
	if err := caps.AllowPath(dir, nono.AccessReadWrite); err != nil {
		panic(err)
	}
	if err := caps.SetNetworkMode(nono.NetworkBlocked); err != nil {
		panic(err)
	}

	state, err := nono.StateFromCaps(caps)
	if err != nil {
		panic(err)
	}
	defer state.Close()

	jsonStr, err := state.JSON()
	if err != nil {
		panic(err)
	}

	restored, err := nono.StateFromJSON(jsonStr)
	if err != nil {
		panic(err)
	}
	defer restored.Close()

	restoredCaps, err := restored.Caps()
	if err != nil {
		panic(err)
	}
	defer restoredCaps.Close()

	fmt.Println(json.Valid([]byte(jsonStr)))
	fmt.Println(restoredCaps.NetworkMode())

	// Output:
	// true
	// blocked
}

func Example_errorHandling() {
	dir, err := os.MkdirTemp("", "nono-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	caps := nono.New()
	defer caps.Close()

	err = caps.AllowPath(filepath.Join(dir, "missing"), nono.AccessRead)
	fmt.Println(errors.Is(err, nono.ErrPathNotFound))

	// Output:
	// true
}
