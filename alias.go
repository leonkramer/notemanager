package main

import (
	"flag"
	"fmt"
	"strings"
)

func aliasHandler(filter NoteFilter, notes []Note, args []string) (err error) {
	/* 	if FilterIsDefined(filter) {
	   		Exit("Can Note ID ")
	   	} else {
	   		Exit("has no Tags")
	   	} */

	var optHelp bool
	fs := flag.NewFlagSet("note alias", flag.ContinueOnError)
	fs.Usage = func() { helpNoteAlias() }
	fs.BoolVar(&optHelp, "h", false, "Display usage")
	fs.BoolVar(&optHelp, "help", false, "Display usage")

	if err = fs.Parse(args); err != nil {
		return
	}

	if optHelp {
		helpNoteAlias()
	}

	args = fs.Args()

	// default: list
	if len(args) == 0 {
		args = append(args, "list")
	}

	cmd, rargs := args[0], args[1:]

	switch cmd {
	case "add":
		if len(notes) > 1 {
			Exit("Adding alias to multiple notes not allowed")
		}

		if len(rargs) == 0 {
			helpNoteAlias()
		}

		n := notes[0]
		n.AddAliases(rargs)

		fmt.Printf("%s: OK\n", n.ShortId())
		n.WriteData()

	case "list":
		for _, n := range notes {
			if len(n.Aliases) > 0 {
				fmt.Printf("%s: %s\n", n.ShortId(), strings.Join(n.Aliases, ", "))
			}
		}

	case "delete":
		if len(notes) > 1 {
			Exit("Deleting aliases of multiple notes not allowed")
		}

		if len(rargs) == 0 {
			helpNoteAlias()
		}

		n := notes[0]
		n.RemoveAliases(rargs)
		fmt.Printf("%s: OK\n", n.ShortId())
		n.WriteData()

	default:
		helpNoteAlias()
	}

	return
}
