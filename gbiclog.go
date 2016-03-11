// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
//
// TASKS:
//DONE: create scheme of DB
//DONE: config - data file name
//DONE: command - init data file
//DONE: checking if given file is a appropriate biclog file
//DONE: command - type add
//DONE: command - type list
//DONE: command - type edit
//DONE: command - type delete
//TODO: command - category add
//TODO: command - category list
//TODO: command - category edit
//TODO: command - category delete
//TODO: command - bicycle add
//TODO: command - bicycle list
//TODO: command - bicycle edit
//TODO: command - bicycle delete
//TODO: command - bicycle show details
//TODO: command - bicycle show photo
//TODO: command - trip add
//TODO: command - trip list
//TODO: command - trip edit
//TODO: command - trip delete
//TODO: command - trip show details
//TODO: command - report summary
//TODO: command - report history
//TODO: command - report pie chart (share of bicycles)
//TODO: command - report bar chart (history)
package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/zbroju/gbiclog/lib/bicycleTypes"
	"github.com/zbroju/gbiclog/lib/sqlitedb"
	"github.com/zbroju/gprops"
	"os"
	"path"
	"strconv"
	"strings"
)

// Error messages
const (
	errMissingFileFlag = "gBicLog: missing information about data file. Specify it with --file or -f flag.\n"
	errMissingNameFlag = "gBicLog: missing name. Specify it with --name or -n flag.\n"
	errMissingIdFlag   = "gBicLog: missing id. Specify it with --id or -i flag.\n"
)

// Config settings
const (
	confDataFile = "DATA_FILE"
	confVerbose  = "VERBOSE"
)

// Formatting strings for display
const (
	fsSeparator = "  "
)

// Objects
const (
	objectBicycleType      = "bicycle_type"
	objectBicycleTypeAlias = "bt"
)

func main() {
	cli.CommandHelpTemplate = `
	NAME:
	   {{.HelpName}} - {{.Usage}}
	USAGE:
	   {{.HelpName}}{{if .Subcommands}} [subcommand]{{end}}{{if .Flags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{if .Description}}
	DESCRIPTION:
	   {{.Description}}{{end}}{{if .Flags}}
	OPTIONS:
	   {{range .Flags}}{{.}}
	   {{end}}{{ end }}{{if .Subcommands}}
	SUBCOMMANDS:
	    {{range .Subcommands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
	{{end}}{{ end }}
	`

	// Loading properties from config file if exists
	configSettings := gprops.New()
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".gblrc"))
	if err == nil {
		err = configSettings.Load(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gBicLog: syntax error in %s. Exit.\n", configFile.Name())
			return
		}
	}
	configFile.Close()
	dataFile := configSettings.GetOrDefault(confDataFile, "")
	verbose, err := strconv.ParseBool(configSettings.GetOrDefault(confVerbose, "false"))
	if err != nil {
		verbose = false
	}

	// Commandline arguments
	app := cli.NewApp()
	app.Name = "gBicLog"
	app.Usage = "keeps track of you bike rides"
	app.Version = "0.1 Alfa"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	// Flags definitions
	flagVerbose := cli.BoolFlag{
		Name:        "verbose, b",
		Usage:       "show more output",
		Destination: &verbose,
	}
	flagFile := cli.StringFlag{
		Name:  "file, f",
		Value: dataFile,
		Usage: "data file",
	}
	flagName := cli.StringFlag{
		Name:  "name, n",
		Value: "",
		Usage: "name",
	}
	flagId := cli.IntFlag{
		Name:  "id, i",
		Value: -1,
		Usage: "ID of an object",
	}

	// Commands
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagVerbose, flagFile},
			Usage:   "Init a new data file specified by the user",
			Action:  cmdInit,
		},
		{
			Name:    "add",
			Aliases: []string{"A"},
			Usage:   "Add an object (bicycle, bicycle type, trip, trip category).",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagName},
					Usage:   "Add new bicycle type.",
					Action:  cmdTypeAdd,
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"L"},
			Usage:   "List objects (bicycles, bicycle types, trips, trips categories)",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile},
					Usage:   "List available bicycle types.",
					Action:  cmdTypeList,
				},
			},
		},
		{
			Name:    "edit",
			Aliases: []string{"E"},
			Usage:   "Edit an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId, flagName},
					Usage:   "Edit bicycle type with given id.",
					Action:  cmdTypeEdit,
				},
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"D"},
			Usage:   "Delete an object (bicycle, bicycle type, trip, trip category)",
			Subcommands: []cli.Command{
				{
					Name:    objectBicycleType,
					Aliases: []string{objectBicycleTypeAlias},
					Flags:   []cli.Flag{flagVerbose, flagFile, flagId},
					Usage:   "Delete bicycle type with given id.",
					Action:  cmdTypeDelete,
				},
			},
		},
	}
	app.Run(os.Args)
}

func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == "" {
		fmt.Fprint(os.Stderr, errMissingFileFlag)
		return
	}

	// Create new file
	f := sqlitedb.New(c.String("file"))
	err := f.CreateNew()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "gBicLog: created file %s.\n", c.String("file"))
	}
}

func cmdTypeAdd(c *cli.Context) {
	// Check obligatory flags (file, name)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, errMissingFileFlag)
		return
	}
	if c.String("name") == "" {
		fmt.Fprintf(os.Stderr, errMissingNameFlag)
		return
	}

	// Open data file
	f := sqlitedb.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	defer f.Close()

	// Add new type
	nt := bicycleTypes.BicycleType{0, c.String("name")}
	err = f.TypeAdd(nt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "gBicLog: added new bicycle type: %s.\n", nt.Name)
	}

}

func cmdTypeList(c *cli.Context) {
	// Check obligatory flags (file)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, errMissingFileFlag)
		return
	}

	// Open data file
	f := sqlitedb.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer f.Close()

	// List bicycle types
	types, err := f.TypeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	if len(types) == 0 {
		fmt.Fprintf(os.Stdout, "gBicLog: no bicycle types.")
		return
	}
	idH, nameH, idFS, nameFS := types.GetDisplayStrings()
	fmt.Fprintf(os.Stdout, strings.Join([]string{idH, nameH, "\n"}, fsSeparator))
	l := strings.Join([]string{idFS, nameFS, "\n"}, fsSeparator)
	for _, t := range types {
		fmt.Fprintf(os.Stdout, l, t.Id, t.Name)
	}
}

func cmdTypeEdit(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, errMissingFileFlag)
		return
	}
	id := c.Int("id")
	if id < 0 {
		fmt.Fprintf(os.Stderr, errMissingIdFlag)
		return
	}
	newName := c.String("name")
	if newName == "" {
		fmt.Fprintf(os.Stderr, errMissingNameFlag)
		return
	}

	// Open data file
	f := sqlitedb.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer f.Close()

	// Edit bicycle type
	btl, err := f.TypeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	bt, err := btl.GetWithId(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	oldName := bt.Name
	bt.Name = newName

	err = f.TypeUpdate(bt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "gBicLog: change bicycle type name from %s to %s.\n", oldName, newName)
	}
}

func cmdTypeDelete(c *cli.Context) {
	// Check obligatory flags
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, errMissingFileFlag)
		return
	}
	id := c.Int("id")
	if id < 0 {
		fmt.Fprintf(os.Stderr, errMissingIdFlag)
		return
	}

	// Open data file
	f := sqlitedb.New(c.String("file"))
	err := f.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer f.Close()

	// Delete bicycle type
	btl, err := f.TypeList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}
	bt, err := btl.GetWithId(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	err = f.TypeDelete(bt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "gBicLog: deleted bicycle type %s.\n", bt.Name)
	}
}
