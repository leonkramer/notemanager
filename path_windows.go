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
	homedir, err := os.UserHomeDir()
	if err != nil {
        log.Fatal(err)
    }

	cfg, err := conf.ReadFile(filepath.Clean(homedir + `/AppData/Roaming/Notemanager/noterc`))
	if err != nil {
        log.Fatal(err)
    }

	datadir, err := cfg.String("default", "datadir")
	if err == nil {
		c.DataDir = datadir
	} else {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		c.DataDir = filepath.Clean(homedir + `/AppData/Roaming/Notemanager`)
	}
	c.TemplateDir = filepath.Clean(c.DataDir + `/templates`)
	c.TempDir = filepath.Clean(c.DataDir + `/tmp`)
	c.NoteDir = filepath.Clean(c.DataDir + `/notes`)
	
	c.VersionTimeFormat = "20060102-150405"
	c.OutputTimeFormatShort = "2006-01-02"
	c.OutputTimeFormatLong = "2006-01-02 15:04:05"

	// not sure why, but explorer.exe always returns exit code 1
	// while 'start' does not.
	c.FileManager = `start /B`

	// File and Directory Permissions
	// Read+Write
	c.FilePermission = 0600
	// ReadOnly. Attachments should be readonly, so they are not being accidently
	// overwritten, when browsing with file manager.
	c.FilePermissionReadonly = 0400
	// Read+Write+Execute
	c.DirPermission = 0600

	c.Editor, err = cfg.String("default", "editor")
	// Default to notepad.exe as editor
	if err != nil {
		c.Editor = `notepad`
	}

	// Default to more.exe as reader with pagination
	c.TerminalReader, err = cfg.String("default", "terminalReader")
	if err != nil {
		c.TerminalReader = `more`
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