package invocation

import (
	"path/filepath"
	"strings"
)

func protonTargetIndex(args []string) int {
	i := protonCommandIndex(args)
	if i < 0 {
		return -1
	}

	return i + 2
}

func protonCommandIndex(args []string) int {
	for i := 0; i+2 < len(args); i++ {
		if filepath.Base(args[i]) == "proton" && args[i+1] == "waitforexitandrun" {
			return i
		}
	}

	return -1
}

// IsProtonCommand reports whether the invocation contains a Proton game command.
func IsProtonCommand(args []string) bool {
	return protonCommandIndex(args) >= 0
}

func ProtonLogFields(args []string) string {
	protonIndex := protonCommandIndex(args)
	if protonIndex < 0 {
		return ""
	}

	proton := filepath.Base(filepath.Dir(args[protonIndex]))
	runtime := ""
	for _, arg := range args[:protonIndex] {
		if filepath.Base(arg) == "_v2-entry-point" {
			runtime = filepath.Base(filepath.Dir(arg))
		}
	}

	var fields strings.Builder
	if runtime != "" {
		fields.WriteString(" runtime=")
		fields.WriteString(runtime)
	}

	if proton != "" && proton != "." {
		fields.WriteString(" proton=")
		fields.WriteString(proton)
	}

	return fields.String()
}
