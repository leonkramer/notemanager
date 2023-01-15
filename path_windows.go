// +build windows

package main

import (
	"os"
	"log"
	"bytes"
	"os/exec"
	"fmt"
	"path/filepath"
	"github.com/gosimple/conf"
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
	c.DirPermission = 0600

	// not sure why, but explorer.exe always returns exit code 1
	// while 'start' does not.
	c.FileManager = `start /B`

	// Pagination Reader (Default: more)
	c.TerminalReader = `more`

	// default data directory
	c.DataDir = filepath.Clean(homedir + `/AppData/Roaming/Notemanager`)

	//c.Editor = `notepad`
	// set default editor
	editors := []string{"nvim", "gvim", "notepad"}
	for _, editor := range editors {
		path, err := exec.LookPath(editor)
		if (err == nil) {
			c.Editor = filepath.Clean(path)
			break
		}
	}

	cfgExists = false
	cfg, err := conf.ReadFile(filepath.Clean(homedir + `/AppData/Roaming/Notemanager/noterc`))
	if err == nil {
		cfgExists = true
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

		// Default to more.exe as reader with pagination
		terminalReader, err := cfg.String("default", "terminalReader")
		if err == nil {
			c.TerminalReader = filepath.Clean(terminalReader)
		}

	}

	c.TemplateDir = filepath.Clean(c.DataDir + `/templates`)
	c.TempDir = filepath.Clean(c.DataDir + `/tmp`)
	c.NoteDir = filepath.Clean(c.DataDir + `/notes`)

	if c.Editor == "" {
		log.Fatal("Please define a text editor")
	}

	return
}

func runFileManager(path string) {
	//cmd := exec.Command("cmd", "/C", notemanager.FileManager, path)
	//command := append([]string{"/C"}, notemanager.FileManager...)
	//command = append(command, path)
	//cmd := exec.Command("cmd", command...)
	cmd := exec.Command("cmd", "/C", notemanager.FileManager, path)


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