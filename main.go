package main

import (
	"fmt"
	"os"
	"log"
	"encoding/hex"
	"regexp"
	"path/filepath"
	"strconv"
	_"time"
	"io"
	"crypto/sha1"
	"github.com/gosimple/conf"
	"gopkg.in/yaml.v3"
)

var cfg *conf.Config
var notemanager Config

func main() {
	notemanager = parseConfig()

	switch os.Args[1] {
		case "list":
			listHandler(os.Args[2:])
		
		case "add":
			addHandler()

/* 		case "read":
			readHandler()

		case "delete":
			deleteHandler() */

		case "help":
			displayUsageGeneric()
		
		case "version":
			fmt.Println(`Notemanager Version 0.15
Author: Leon Kramer <leonkramer@gmail.com>`)

		default:
			noteHandler()
	}
}

func displayUsageGeneric() {
	fmt.Println(`Notemanager Usage:
	
Generic usage:
---
note add [ +TAG .. ] TITLE
	Add note
note help
	Display usage
note list [ OPTIONS ] [ FILTER ]
	List notes which match filter
	Options:
		-a		List all notes, include deleted
note version
	Display version


Note specific usage:
---
note ID [ read ]
	Read note with pagination
note ID print
	Print note
note ID delete
	Mark note as deleted
`)
}

// moves temporary note from tempDir to specific note directory inside noteDir
func moveFile(id string, version string) (err error) {
	oldFile := fmt.Sprintf("%s/%s", notemanager.TempDir, id)
	newFile := fmt.Sprintf("%s/%s/%s", notemanager.NoteDir, id, version)

	os.Mkdir(filepath.Clean(fmt.Sprintf("%s/%s", notemanager.NoteDir, id)), notemanager.FilePermission)
	err = os.Rename(oldFile, newFile)

	return
}

func readMetadataFile(id string) (metadata Metadata, err error) {
	metadataRaw, err := os.ReadFile(filepath.Clean(notemanager.NoteDir + "/" + id + "/meta"))
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


// generate sha1 hash from file
func fileSha1(path string) (ret string, err error) {
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer fh.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, fh)
	if err != nil {
		return
	}
	hashInBytes := hash.Sum(nil)[:20]
	ret = hex.EncodeToString(hashInBytes)
	return
}

