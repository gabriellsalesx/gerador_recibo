package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenFile(path string) error {
	if path == "" {
		return fmt.Errorf("caminho do arquivo não informado")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("arquivo não encontrado: %s", abs)
		}
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", abs)
	case "darwin":
		cmd = exec.Command("open", abs)
	default:
		cmd = exec.Command("xdg-open", abs)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("não consegui abrir o arquivo no visualizador padrão: %w", err)
	}
	return nil
}
