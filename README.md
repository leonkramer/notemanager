# Notemanager
A lightweight cross platform note manager for CLI.

Are you tired of losing track of your digital note taking? Notemanager is a lightweight software which integrates into your terminal and organizes your notes. 
The syntax is inspired by Taskwarrior, a popular task manager for the CLI. Notemanager lets you use your favorite editor for note taking. As it is written in Go, Notemanager can be compiled and used by hackers using Linux, MacOS or Windows.


## Further Reads
For further information, please visit the wiki @ https://github.com/leonkramer/notemanager/wiki


### Todo: Move -> Wiki
# Templates
Create text templates in your template directory. Templates can be used for note creation and allow to use placeholders. Placeholders are wrapped in double curly braces `{{ }}`. The following placeholders will be replaced with note specific data:
* `nmId`
* `nmTitle`
* `nmTags`
* `nmDate`
* `nmTime`
* `nmTimeOffset`

The following template is an example of a specific template layout. Assuming you name this file _weekly_, you can use the parameter `-t weekly` with the `note add` command to use the template.
```
---
Weekly Meeting
Participants:   
Date:           {{ nmDate }}
---

# Weekly Meeting @ {{ nmDate }}

## Topics
1.

## Todos
1. 

```
