package main

import (
	"flag"
	"fmt"
	"log"
)

func editHandler(notes []Note, args []string) (err error) {
	if len(notes) > 1 {
		Exit("Only supply one note")
	}

	if len(args) > 0 {
		Exit("Error")
	}
	note := notes[0]

	err = noteEditHandler(note)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func aliasHandler(filter NoteFilter, notes []Note, args []string) (err error) {
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

	if len(args) == 0 {
		helpNoteAlias()
	}

	cmd, rargs := args[0], args[1:]

	switch cmd {
	case "set":
		err = aliasSetHandler(notes, rargs)
		if err != nil {
			Exit(err.Error())
		}

	case "list":
		for _, n := range notes {
			fmt.Printf("%s: %s\n", n.ShortId(), n.Alias)
		}

	case "remove":
		if len(notes) > 1 {
			helpNoteAlias()
		}

		if len(rargs) > 0 {
			helpNoteAlias()
		}

		n := notes[0]
		n.Alias = ""

		aliases.DeleteById(n.Id)
		aliases.Write()

		fmt.Printf("%s: OK\n", n.ShortId())
		n.WriteData()

	default:
		helpNoteAlias()
	}

	return
}

func aliasSetHandler(notes []Note, args []string) (err error) {
	if len(args) == 0 || len(args) > 1 {
		helpNoteAlias()
	}

	if len(notes) > 1 {
		helpNoteAlias()
	}

	n := notes[0]
	n.Alias = args[0]

	aliases.DeleteById(n.Id)
	aliases.Set(n.Alias, n.Id)
	aliases.Write()

	fmt.Printf("%s: OK\n", n.ShortId())
	n.WriteData()

	return
}
