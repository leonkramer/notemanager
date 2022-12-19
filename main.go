package main

import (
	"fmt"
	"os"
	"log"
	"regexp"
	"strconv"
	_"time"
	"github.com/gosimple/conf"
	"gopkg.in/yaml.v3"
)

var cfg *conf.Config
var notemanager Config

func main() {
	notemanager = parseConfig()

	switch os.Args[1] {
		case "list":
			listHandler()
		
		case "add":
			addHandler()

/* 		case "read":
			readHandler()

		case "delete":
			deleteHandler() */

		case "help":
			displayUsageGeneric()
		
		case "version":
			fmt.Println(`Notemanager Version 0.1
Author: Leon Kramer <leonkramer@gmail.com>`)

		default:
			noteHandler()
	}
}

func displayUsageGeneric() {
	fmt.Println(`Notemanager Usage:
	
	Generic usage:
	---
	add [ +TAG .. ] TITLE		Add note
	help						Display usage
	list [ FILTER ] 			List notes which match filter
	version						Display version

	Note specific usage:
	---
	ID [ read ]					Read note with pagination
	ID print					Print note
	ID delete					Mark note as deleted
	`)
}

// moves temporary note from tempDir to specific note directory inside noteDir
func moveFile(id string, version string) (err error) {
	oldFile := fmt.Sprintf("%s/%s", notemanager.TempDir, id)
	newFile := fmt.Sprintf("%s/%s/%s", notemanager.NoteDir, id, version)

	os.Mkdir(fmt.Sprintf("%s/%s", notemanager.NoteDir, id), os.ModePerm)
	err = os.Rename(oldFile, newFile)

	return
}

func readMetadataFile(id string) (metadata Metadata, err error) {
	metadataRaw, err := os.ReadFile(notemanager.NoteDir + "/" + id + "/meta")
	if err != nil {
		return
	}

	err = yaml.Unmarshal(metadataRaw, &metadata)
	return metadata, err
}


func noteId(file string) []byte {
	body, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	rUuid := regexp.MustCompile(`(?m)^\-\-\-$[\d\D]+^Id: (?P<uuid>[a-z0-9\-]+)$[\d\D]+^\-\-\-$`)
	rVersion := regexp.MustCompile(`(?m)^\-\-\-$[\d\D]+^Version: (?P<version>[0-9]+)$[\d\D]+^\-\-\-$`)
	if rUuid.Match(body) && rVersion.Match(body) {
		matches := rUuid.FindSubmatch(body)
		uuid := matches[rUuid.SubexpIndex("uuid")]
		matches = rVersion.FindSubmatch(body)
		version := matches[rVersion.SubexpIndex("version")]
		versionNew, err := strconv.Atoi(string(version))
		versionNew += 1
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Uuid: %s", uuid)
		fmt.Printf("Version: %s", version)
		fmt.Println("VersionNew: ", versionNew)
		datadir, _ := cfg.String("default", "datadir")
		dstDir := fmt.Sprintf("%s/%s", datadir, uuid)
		os.MkdirAll(dstDir, 0655)
		if _, err := os.Stat(dstDir); os.IsNotExist(err) {
			fmt.Println("directory does not exist")
		}
	}
	
	return body
}


