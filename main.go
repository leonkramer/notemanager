package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gosimple/conf"
)

var cfg *conf.Config
var notemanager Config

func main() {
	// disable timestamp in Fatal Logs
	log.SetFlags(0)

	// return exit code 1 after runtime.Goexit() function
	// used in func Exit(string).
	// If runtime finishes normally, we need to manually
	// exit with os.Exit(0).
	defer os.Exit(1)

	notemanager = parseConfig()

	var optHelp bool
	var optAll bool
	fs := flag.NewFlagSet("note", flag.ContinueOnError)
	fs.Usage = func() { helpNote() }
	fs.BoolVar(&optAll, "a", false, "Select all notes in filter, include deleted")
	fs.BoolVar(&optAll, "all", false, "Select all notes in filter, include deleted")
	fs.BoolVar(&optHelp, "h", false, "Display usage")
	fs.BoolVar(&optHelp, "help", false, "Display usage")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return
	}

	if optHelp {
		helpNote()
	}

	// remaining args
	rargs := fs.Args()

	// expected cmd syntax: ./note [ FILTER ] cmd args
	filter, rargs, err := parseFilter(rargs)
	if err != nil {
		Exit(`Failed to parse arguments`)
	}

	if optAll {
		filter.IncludeDeleted = true
	}

	if len(rargs) == 0 {
		helpNote()
	}

	notes, err := notes(filter)

	/*
		Handlers which pass the filter should be cleaned up, because we
		built the note selection already. No need to process and pass the
		filter, we can already pass the note selection, i.e. the []Note
		slice.
		Out commented handlers need to be created.
	*/
	switch rargs[0] {
	case "add":
		addHandler(rargs[1:])

	case "delete":
		deleteHandler(notes, rargs[1:])

	//case "edit":
	//	editHandler(notes, rargs[1:])

	//case "file":
	//	fileHandler(notes, rargs[1:])

	case "list":
		listHandler(filter, rargs[1:])

	//case "modify":
	//	modifyHandler(notes, rargs[1:])

	case "print":
		printHandler(notes, rargs[1:])

	//case "purge":
	//	purgeHandler(notes, rargs[1:])

	case "read":
		readHandler(notes, rargs[1:])

	case "search":
		searchHandler(filter, rargs[1:])

	case "tags":
		tagsHandler(filter, rargs[1:])

	case "undelete":
		undeleteHandler(notes, rargs[1:])

	case "version":
		fmt.Println(`Notemanager Version 0.60.1-alpha
Author: Leon Kramer <leonkramer@gmail.com>`)

	default:
		// this handler should be replaced by
		// other specific handlers
		noteHandler()
	}

	// runtime ended properly, so exit with code 0.
	// this is necessary as we deferred os.Exit(1) initially,
	// so runtime.Goexit() returns a proper exit code.
	os.Exit(0)
}
