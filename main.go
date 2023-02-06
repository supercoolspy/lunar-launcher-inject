package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func GetLunarExecutable() (string, error) {
	var exe string

	if len(os.Args) > 1 {
		exe = os.Args[1]
	} else {
		switch runtime.GOOS {
		case "windows":
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			exe = home + `\AppData\Local\Programs\lunarclient\Lunar Client.exe`
		case "darwin":
			exe = "/Applications/Lunar Client.app/Contents/MacOS/Lunar Client"
		case "linux":
			exe = "/usr/bin/lunarclient"
		default:
			return "", fmt.Errorf("locating lunar is not supported on %s", runtime.GOOS)
		}
	}

	if _, err := os.Stat(exe); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("'%s' does not exist", exe)
	}

	return exe, nil
}

func Run() (err error) {
	lunarExe, err := GetLunarExecutable()
	if err != nil {
		if len(os.Args) < 2 {
			return fmt.Errorf("failed to locate the lunar launcher executable, try passing it by argument: %w", err)
		}
		return err
	}

	d, cmd, err := StartProcessAndConnectDebugger(lunarExe)
	defer func() {
		if cmd != nil && err != nil {
			_ = cmd.Process.Kill()
		}
	}()
	if err != nil {
		return err
	}
	defer d.Close()

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	return d.Send("Runtime.callFunctionOn", map[string]any{
		"executionContextId":  1,
		"functionDeclaration": injectJs,
		"arguments": []any{
			map[string]any{
				"value": filepath.Dir(ex),
			},
		},
	})
}

func main() {
	log.SetFlags(0)

	if err := Run(); err != nil {
		log.Fatalln(err)
	}
}

//go:embed inject.js
var injectJs string
