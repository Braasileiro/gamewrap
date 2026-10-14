package host

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type commandCall struct {
	Name string
	Args []string
}

type fakeCommandResult struct {
	Output []byte
	Error  error
}

type fakeCommandRunner struct {
	Results map[string]fakeCommandResult
	Calls   []commandCall
}

func (f *fakeCommandRunner) Run(name string, args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, commandCall{Name: name, Args: append([]string(nil), args...)})
	if result, found := f.Results[commandKey(name, args...)]; found {
		return result.Output, result.Error
	}

	return nil, nil
}

func commandKey(name string, args ...string) string {
	return fmt.Sprintf("%s\x00%s", name, strings.Join(args, "\x00"))
}

func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}
