package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"emissor/internal/core"
)

type FileSet struct {
	Directory  string
	BaseName   string
	PDFPath    string
	JSONPath   string
	PixPNGPath string
}

// BuildFileSet monta os caminhos de saída de um documento, organizados por tipo
// e por data: <baseDir>/<Tipo>/YYYY/MM/DD/<prefixo>-<timestamp>-<numero>.
func BuildFileSet(baseDir string, docType core.DocType, createdAt time.Time, number string) (FileSet, error) {
	baseDir, err := ExpandPath(baseDir)
	if err != nil {
		return FileSet{}, err
	}
	dir := filepath.Join(baseDir, docType.Folder(),
		createdAt.Format("2006"), createdAt.Format("01"), createdAt.Format("02"))
	name := fmt.Sprintf("%s-%s-%s", docType.FilePrefix(), createdAt.Format("20060102-150405"), number)
	return FileSet{
		Directory:  dir,
		BaseName:   name,
		PDFPath:    filepath.Join(dir, name+".pdf"),
		JSONPath:   filepath.Join(dir, name+".json"),
		PixPNGPath: filepath.Join(dir, name+"-pix.png"),
	}, nil
}

func FormatNumber(next, padding int, prefix string) string {
	if padding <= 0 {
		padding = 4
	}
	return fmt.Sprintf("%s%0*d", prefix, padding, next)
}

func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
