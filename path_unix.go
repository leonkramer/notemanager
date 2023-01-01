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
)

func parseConfig() (c Config) {
	homedir, err := os.UserHomeDir()
	if err != nil {
        log.Fatal(err)
    }

	cfg, err := conf.ReadFile(homedir + "/.noterc")
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
		c.DataDir = homedir + "/.notes"
	}
	c.TemplateDir = c.DataDir + "/templates"
	c.TempDir = c.DataDir + "/tmp"
	c.NoteDir = c.DataDir + "/notes"
	c.FileManager = "Open"
	c.VersionTimeFormat = "20060102-150405"
	c.OutputTimeFormatShort = "2006-01-02"
	c.OutputTimeFormatLong = "2006-01-02 15:04:05"

	// File and Directory Permissions

	// Read+Write
	c.FilePermission = 0600
	// ReadOnly. Attachments should be readonly, so they are not being accidently
	// overwritten, when browsing with file manager.
	c.FilePermissionReadonly = 0400
	// Read+Write+Execute
	c.DirPermission = 0700

	editor, err := cfg.String("default", "editor")
	if err == nil {
		c.Editor = editor
	} else {
		log.Fatal("Mandatory setting 'editor' is missing")
	}

	c.TerminalReader, err = cfg.String("default", "terminalReader")
	if err != nil {
		log.Fatal("Mandatory setting 'terminalReader' is missing")
	}

	return
}


func runFileManager(path string) {
	path = filepath.Clean(path)
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