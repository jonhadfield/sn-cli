package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"syscall"
	"time"

	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

const (
	snAppName                          = "sn-cli"
	msgAddSuccess                      = "added"
	msgDeleted                         = "deleted"
	msgMultipleNotesFoundWithSameTitle = "multiple notes found with the same title"
	msgNoteAdded                       = "note added"
	msgNoteDeleted                     = "note deleted"
	msgNoteNotFound                    = "note not found"
	msgTagSuccess                      = "item tagged"
	msgTagAlreadyExists                = "tag already exists"
	msgTagAdded                        = "tag added"
	msgTagDeleted                      = "tag deleted"
	msgFailedToDeleteTag               = "failed to delete tag"
	msgTagNotFound                     = "tag not found"
	msgItemsDeleted                    = "items deleted"
	msgNoMatches                       = "no matches"
	msgRegisterSuccess                 = "registered"
)

var yamlAbbrevs = []string{"yml", "yaml"}

// overwritten at build time.
var version, versionOutput, tag, sha, buildDate string

func main() {
	if err := startCLI(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

type configOptsOutput struct {
	useStdOut  bool
	useSession bool
	sessKey    string
	server     string
	cacheDBDir string
	debug      bool
}

func getOpts(c *cli.Context) (out configOptsOutput) {
	out.useStdOut = true

	if c.Bool("no-stdout") {
		out.useStdOut = false
	}

	if c.Bool("use-session") || viper.GetBool("use_session") {
		out.useSession = true
	}

	out.sessKey = c.String("session-key")

	out.server = c.String("server")

	if viper.GetString("server") != "" {
		out.server = viper.GetString("server")
	}

	out.cacheDBDir = viper.GetString("cachedb_dir")
	if out.cacheDBDir != "" {
		out.cacheDBDir = c.String("cachedb-dir")
	}

	if c.Bool("debug") {
		out.debug = true
	}

	return
}

func appSetup() (app *cli.App) {
	viper.SetEnvPrefix("sn")
	viper.AutomaticEnv()

	if tag != "" && buildDate != "" {
		versionOutput = fmt.Sprintf("[%s-%s] %s", tag, sha, buildDate)
	} else {
		versionOutput = version
	}

	app = cli.NewApp()
	app.EnableBashCompletion = true

	app.Name = "sn"
	app.Version = versionOutput
	app.Compiled = time.Now()
	app.Authors = []*cli.Author{
		{
			Name:  "Jon Hadfield",
			Email: "jon@lessknown.co.uk",
		},
	}
	app.Usage = "Standard Notes CLI"
	app.Description = ""
	app.BashComplete = func(c *cli.Context) {
		for _, cmd := range c.App.Commands {
			if !cmd.Hidden {
				fmt.Fprintln(c.App.Writer, cmd.Name)
			}
		}
	}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{Name: "debug", Value: viper.GetBool("debug")},
		&cli.StringFlag{Name: "server", Value: viper.GetString("server")},
		&cli.BoolFlag{Name: "use-session", Value: viper.GetBool("use_session")},
		&cli.StringFlag{Name: "session-key"},
		&cli.StringFlag{
			Name:  "session-file",
			Usage: "store the session in this file instead of the system keyring, e.g. on a headless server",
			Value: viper.GetString("session_file"),
		},
		&cli.BoolFlag{Name: "no-stdout", Hidden: true},
		&cli.StringFlag{Name: "cachedb-dir", Value: viper.GetString("cachedb_dir")},
	}
	app.Commands = []*cli.Command{
		cmdAdd(),
		cmdBackup(),
		cmdDebug(),
		cmdDelete(),
		cmdEdit(),
		cmdEditor(),
		cmdExport(),
		cmdGet(),
		cmdHealthcheck(),
		cmdMigrate(),
		cmdOrganize(),
		cmdRegister(),
		cmdResync(),
		cmdSearch(),
		cmdSession(),
		cmdStats(),
		cmdTask(),
		cmdTag(),
		cmdTemplate(),
		cmdWipe(),
	}

	app.Before = func(c *cli.Context) error {
		return useSessionFile(c.String("session-file"))
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		_, _ = fmt.Fprintf(c.App.Writer, "\ninvalid command: \"%s\" \n\n", command)

		// Show help for wherever the user actually was. An unknown subcommand
		// of "get" should list what "get" accepts, not the top-level commands.
		if inSubcommand(c) {
			cli.ShowSubcommandHelpAndExit(c, 1)
		}

		cli.ShowAppHelpAndExit(c, 1)
	}

	return app
}

// inSubcommand reports whether the context belongs to a subcommand rather than
// the application itself. urfave/cli wraps the app in a synthetic root command
// named after the app, so that name is what distinguishes the two.
func inSubcommand(c *cli.Context) bool {
	if c == nil || c.Command == nil {
		return false
	}

	return c.Command.Name != "" && c.Command.Name != c.App.Name
}

func startCLI(args []string) (err error) {
	app := appSetup()

	sort.Sort(cli.FlagsByName(app.Flags))

	return app.Run(args)
}

func getPassword() (res string, err error) {
	for {
		fmt.Print("password: ")
		var bytePassword []byte
		if bytePassword, err = term.ReadPassword(int(syscall.Stdin)); err != nil {
			return
		}

		if len(bytePassword) < sncli.MinPasswordLength {
			err = fmt.Errorf("\rpassword must be at least %d characters", sncli.MinPasswordLength)

			return
		}

		var bytePassword2 []byte
		fmt.Printf("\rconfirm password: ")
		if bytePassword2, err = term.ReadPassword(int(syscall.Stdin)); err != nil {
			return
		}

		if !bytes.Equal(bytePassword, bytePassword2) {
			fmt.Printf("\rpasswords do not match")
			fmt.Println()

			return
		}

		fmt.Println()
		if err == nil {
			res = string(bytePassword)

			return
		}
	}
}

func numTrue(in ...bool) (total int) {
	for _, i := range in {
		if i {
			total++
		}
	}

	return
}
