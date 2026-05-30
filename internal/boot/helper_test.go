package boot

import (
	"os"
	"path/filepath"
	"strings"
)

// writeFileRaw writes body to dir/name, trimming a leading newline so test
// literals can start on the line after the backtick.
func writeFileRaw(dir, name, body string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(strings.TrimPrefix(body, "\n")), 0o644)
}
