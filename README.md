# Notemanager
A lightweight note manager for CLI.

Are you tired of losing track of your digital note taking? Notemanager is a lightweight software which integrates into your terminal. The syntax is inspired by Taskwarrior, a heavily used CLItask manager.


# Installation

## Config
```
# ~/.noterc
datadir = /Users/leon/.notes
editor = /usr/bin/vim
terminalReader = /usr/bin/less

```

## Create Folders
```
mkdir -p ~/.notes/{notes,templates,tmp}
```


# Usage

```
	Generic usage:
	---
	add [ +TAG .. ] TITLE		Add note
	help						        Display usage
	list [ FILTER ] 		  	List notes which match filter
	version						      Display version


	Note specific usage:
	---
	ID [ read ]					    Read note with pagination
	ID print					      Print note
	ID delete					      Mark note as deleted
```

