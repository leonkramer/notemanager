// +build !windows

package main

import (
	"os"
	"log"
	"os/exec"
	"bytes"
	"fmt"
	"path/filepath"
	"github.com/gosimple/conf"
	"golang.org/x/term"
)

func parseConfig() (c Config) {
	var cfgExists bool

	homedir, err := os.UserHomeDir()
	if err != nil {
        log.Fatal(err)
    }

	// Syntax of note version file names
	c.VersionTimeFormat = "20060102-150405"

	// Short timestamp format
	c.OutputTimeFormatShort = "2006-01-02"

	// Long timestamp format
	c.OutputTimeFormatLong = "2006-01-02 15:04:05"

	// File and Directory Permissions
	// Read+Write
	c.FilePermission = 0600

	// ReadOnly. Attachments should be readonly, so they are not being accidently
	// overwritten, when browsing with file manager.
	c.FilePermissionReadonly = 0400

	// Read+Write+Execute
	c.DirPermission = 0700

	// Command Wrapper for Default Applications
	c.FileManager = `open`

	// Pagination Reader (Default: less)
	c.TerminalReader = `less`

	// default data directory
	c.DataDir = filepath.Clean(homedir + "/.notes")

	// set default editor
	editors := []string{"nano", "nvim", "vim", "vi", "emacs", "ed"}
	for _, editor := range editors {
		path, err := exec.LookPath(editor)
		if (err == nil) {
			c.Editor = filepath.Clean(path)
			break
		}
	}
	
	cfg, err := conf.ReadFile(filepath.Clean(homedir + "/.noterc"))
	if err == nil {
		cfgExists = true
    } else {
		cfgExists = false
	}

	if cfgExists {
		datadir, err := cfg.String("default", "datadir")
		if err == nil {
			c.DataDir = filepath.Clean(datadir)
		}

		editor, err := cfg.String("default", "editor")
		if err == nil {
			c.Editor = filepath.Clean(editor)
		}

		terminalReader, err := cfg.String("default", "terminalReader")
		if err == nil {
			c.TerminalReader = filepath.Clean(terminalReader)
		}
	}

	c.TemplateDir = filepath.Clean(c.DataDir + "/templates")
	c.TempDir = filepath.Clean(c.DataDir + "/tmp")
	c.NoteDir = filepath.Clean(c.DataDir + "/notes")

	if c.Editor == "" {
		log.Fatal("Please define a text editor")
	}

	return
}


func runFileManager(path string) {
	//command := append([]Any{"cmd", "/C"}, notemanager.FileManager..., path)
	//command := append(notemanager.FileManager, path)
	cmd := exec.Command(notemanager.FileManager, path)
	
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
}

func terminalWidth() (width int) {
	width, _, _ = term.GetSize(0)
	return
}