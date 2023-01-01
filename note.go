package main

import (
	"errors"
	"os"
	"io"
	"log"
	"fmt"
	"regexp"
	_"time"
	"path/filepath"
	"github.com/google/uuid"
)



// Generate a uuid from a abbreviated uuid string.
// The abbreviated string must be at the start of the uuid and be
// 8 bytes long.
// In order to get the real uuid, it must exist in note directory.
// Example: 1cf77aeb => 1cf77aeb-fcb2-44ad-87d6-69717dba1d0c
func uuidByAbbr(x string) (r uuid.UUID, err error) {
	if isUuidAbbr(x) == false {
		//err = errors.New("%s: No such note id")
		err = fmt.Errorf("%s: No such note id", r.String())
		return
	}
	
	matches, err := filepath.Glob(notemanager.NoteDir + "/" + x + "*")
	if (err != nil) {
		return
	}

	switch len(matches) {
	case 0:
		err = errors.New("No such note id")

	case 1:
		r, err = uuid.Parse(filepath.Base(matches[0]))

	default:
		err = errors.New("At least 2 notes found starting with " + x + ", use full uuid")
	}

	return
}

// Checks if input string is a valid Uuid abbreviation.
// Should be syntax: [a-f0-9]{8}
func isUuidAbbr(x string) (bool) {
	reg := regexp.MustCompile(`^[a-f0-9]{8}$`)
	if reg.MatchString(x) {
		return true
	}
	return false
}


// Copies a regular file from src path to dst path.
// if dst already exists return an error.
func copyRegularFile(src string, dst string) (err error) {
	sfh, err := os.Stat(src)
    if err != nil {
        return
    }

	if sfh.Mode().IsRegular() == false {
		err = errors.New("Not a regular file: " + src)
		return
	}

	// check if file exists
	_, err = os.Stat(dst)
	if err == nil {
		err = errors.New("File already exists. Aborting.")
		return err
	}

	sf, err := os.Open(src)
	if err != nil {
        return
    }
	defer sf.Close()


	//df, err := os.Create(dst)
	//df, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL, notemanager.FilePermissionReadonly)
	df, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, notemanager.FilePermissionReadonly)
	if err != nil {
        return
    }
	defer df.Close()
	
	// copy file
	n, err := io.Copy(df, sf)
	if err != nil {
		log.Fatal("n:", n, "err:", err)
	}
	// compare checksums
	srcHash, err := fileSha1(src)
	if err != nil {
		log.Fatal(err)
	}
	dstHash, err := fileSha1(dst)
	if err != nil {
		log.Fatal(err)
	}
	if srcHash == dstHash {
		// files are same, OK
		return
	} else {
		err = errors.New("Copy failed")
		return
	}
}
