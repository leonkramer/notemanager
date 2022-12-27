// +build windows

package main

import (
	"os"
	_"fmt"
	"log"
	"github.com/gosimple/conf"
)

func parseConfig() (c Config) {
	homedir, err := os.UserHomeDir()
	if err != nil {
        log.Fatal(err)
    }

	cfg, err := conf.ReadFile(homedir + `/AppData/Roaming/Notemanager/noterc`)
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
		c.DataDir = homedir + `/AppData/Roaming/Notemanager`
	}
	c.TemplateDir = c.DataDir + `/templates`
	c.TempDir = c.DataDir + `/tmp`
	c.NoteDir = c.DataDir + `/notes`
	c.VersionTimeFormat = "20060102-150405"
	c.OutputTimeFormatShort = "2006-01-02"
	c.OutputTimeFormatLong = "2006-01-02 15:04:05"

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
