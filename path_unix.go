// +build !windows

package main

import (
	"os"
	"log"
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
	c.VersionTimeFormat = "20060102-150405"
	c.OutputTimeFormatShort = "2006-01-02"
	c.OutputTimeFormatLong = "2006-01-02 15:04:05"

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