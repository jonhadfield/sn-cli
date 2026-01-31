package main

import (
	"fmt"
	"slices"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const (
	txtDefaultGroup        = "default_group"
	txtDefaultList         = "default_list"
	msgTaskAdded           = "task added"
	msgTaskCompleted       = "task completed"
	msgTaskReopened        = "task reopened"
	msgTaskDeleted         = "task deleted"
	msgGroupAdded          = "group added"
	msgGroupDeleted        = "group deleted"
	txtOrdering            = "ordering"
	txtOrderingLastUpdated = "last-updated"
	txtOrderingStandard    = "standard"
	txtCompleted           = "completed"
	flagTasklistName       = "list"
	flagTitleName          = "title"
	flagTaskName           = "task"
	flagGroupName          = "group"
	flagUUIDName           = "uuid"
	defaultShowCompleted   = false
)

func cmdTask() *cli.Command {
	return &cli.Command{
		Name:  "task",
		Usage: "manage checklist tasks",
		BashComplete: func(c *cli.Context) {
			addTasks := []string{"add", "list", "show", "complete", "reopen", "delete"}
			if c.NArg() > 0 {
				return
			}
			for _, t := range addTasks {
				fmt.Println(t)
			}
		},
		Subcommands: []*cli.Command{
			cmdTaskAddTask(),
			cmdTaskComplete(),
			cmdTaskDelete(),
			cmdTaskList(),
			cmdTaskReopen(),
			cmdTaskShow(),
		},
	}
}

func cmdTaskList() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "list",
		Subcommands: nil,
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}
			listTasklistsConfig := sncli.ListTasklistsInput{
				Session: &sess,
				Debug:   c.Bool("debug"),
			}

			if err = listTasklistsConfig.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}

func cmdTaskAddTask() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "create a new task",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagTasklistName, Aliases: []string{"l"}, Value: viper.GetString(txtDefaultList)},
			&cli.StringFlag{Name: flagGroupName, Aliases: []string{"g"}, Value: viper.GetString(txtDefaultGroup)},
			&cli.StringFlag{Name: flagTitleName, Required: true},
		},
		BashComplete: func(c *cli.Context) {
			addTasks := []string{"--title", "--list", "--group"}
			if c.NArg() > 0 {
				return
			}
			for _, t := range addTasks {
				fmt.Println(t)
			}
		},
		Action: func(c *cli.Context) error {
			if c.String(flagTasklistName) == "" && c.String(flagUUIDName) == "" {
				return fmt.Errorf("either --%s or --%s must be specified", flagTasklistName, flagUUIDName)
			}

			opts := getOpts(c)

			var sess cache.Session
			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			if c.String(flagGroupName) != "" {
				addTaskInput := sncli.AddAdvancedChecklistTaskInput{
					Session:  &sess,
					Tasklist: c.String(flagTasklistName),
					Group:    c.String(flagGroupName),
					Title:    c.String(flagTitleName),
				}

				if err = addTaskInput.Run(); err != nil {
					return err
				}
			} else {
				addTaskInput := sncli.AddTaskInput{
					Session:  &sess,
					Tasklist: c.String(flagTasklistName),
					Title:    c.String(flagTitleName),
					UUID:     c.String(flagUUIDName),
				}
				if err = addTaskInput.Run(); err != nil {
					return err
				}
			}

			fmt.Println(msgTaskAdded)

			return nil
		},
	}
}

func cmdTaskShow() *cli.Command {
	viper.SetDefault("show_completed", defaultShowCompleted)

	return &cli.Command{
		Name:  "show",
		Usage: "show list",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagTasklistName, Aliases: []string{"l"}, Value: viper.GetString(txtDefaultList)},
			&cli.StringFlag{Name: flagUUIDName},
			&cli.BoolFlag{Name: txtCompleted, Aliases: []string{"c"}, Value: viper.GetBool("show_completed"), Usage: "show completed tasks"},
			&cli.StringFlag{
				Name:    txtOrdering,
				Aliases: []string{"order"},
				Hidden:  true,
				Usage:   "order by standard (as shown in the app) or last-updated",
				Value:   txtOrderingStandard,
			},
		},

		Action: func(c *cli.Context) error {
			var err error

			ordering := c.String(txtOrdering)
			if !slices.Contains([]string{txtOrderingStandard, txtOrderingLastUpdated}, ordering) {
				return fmt.Errorf("%s must be either standard or last-updated", txtOrdering)
			}

			opts := getOpts(c)

			var sess cache.Session
			sess, _, err = cache.GetSession(common.NewHTTPClient(), viper.GetBool("use_session"), opts.sessKey, viper.GetString("server"), opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			showTasklistInput := sncli.ShowTasklistInput{
				Session:       &sess,
				Title:         c.String(flagTasklistName),
				UUID:          c.String(flagUUIDName),
				ShowCompleted: c.Bool(txtCompleted),
				Ordering:      ordering,
			}

			return showTasklistInput.Run()
		},
	}
}

func cmdTaskDelete() *cli.Command {
	return &cli.Command{
		Name:  "delete-task",
		Usage: "delete task",
		Aliases: []string{
			"delete",
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagTasklistName, Aliases: []string{"l"}, Value: viper.GetString(txtDefaultList)},
			&cli.StringFlag{Name: flagGroupName, Aliases: []string{"g"}, Value: viper.GetString(txtDefaultGroup)},
			&cli.StringFlag{Name: flagTitleName, Aliases: []string{flagTaskName}},
		},
		BashComplete: func(c *cli.Context) {
			addTasks := []string{"--title", "--list", "--group"}
			if c.NArg() > 0 {
				return
			}
			for _, t := range addTasks {
				fmt.Println(t)
			}
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			group := c.String(flagGroupName)
			if group != "" {
				deleteTaskInput := sncli.DeleteAdvancedChecklistTaskInput{
					Session:   &sess,
					Debug:     c.Bool("debug"),
					Title:     c.String(flagTitleName),
					Group:     c.String(flagGroupName),
					Checklist: c.String(flagTasklistName),
				}
				err = deleteTaskInput.Run()
				if err != nil {
					return err
				}
			} else {
				deleteTaskInput := sncli.DeleteTaskInput{
					Session:  &sess,
					Debug:    c.Bool("debug"),
					Tasklist: c.String(flagTasklistName),
					Title:    c.String(flagTitleName),
				}

				if err = deleteTaskInput.Run(); err != nil {
					return err
				}
			}

			fmt.Println(msgTaskDeleted)

			return nil
		},
	}
}

func cmdTaskComplete() *cli.Command {
	return &cli.Command{
		Name:  "complete",
		Usage: "complete task",
		Aliases: []string{
			"close",
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagTasklistName, Aliases: []string{"l"}, Value: viper.GetString(txtDefaultList)},
			&cli.StringFlag{Name: flagGroupName, Aliases: []string{"g"}, Value: viper.GetString(txtDefaultGroup)},
			&cli.StringFlag{Name: flagTitleName},
			&cli.StringFlag{Name: flagUUIDName},
		},
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}
			for _, t := range []string{"--title", "--list", "--group", "--uuid"} {
				fmt.Println(t)
			}
		},
		Action: func(c *cli.Context) error {
			if c.String(flagTasklistName) == "" && c.String(flagUUIDName) == "" {
				return fmt.Errorf("either --%s or --%s must be specified", flagTasklistName, flagUUIDName)
			}

			opts := getOpts(c)

			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			if c.String(flagGroupName) != "" {
				completeAdvancedTaskInput := sncli.CompleteAdvancedTaskInput{
					Session:  &sess,
					Tasklist: c.String(flagTasklistName),
					Group:    c.String(flagGroupName),
					Title:    c.String(flagTitleName),
				}

				if err = completeAdvancedTaskInput.Run(); err != nil {
					return err
				}
			} else {
				completeTaskInput := sncli.CompleteTaskInput{
					Session:  &sess,
					Tasklist: c.String(flagTasklistName),
					Title:    c.String(flagTitleName),
				}

				if err = completeTaskInput.Run(); err != nil {
					return err
				}
			}

			fmt.Println(msgTaskCompleted)

			return nil
		},
	}
}

func cmdTaskReopen() *cli.Command {
	return &cli.Command{
		Name:  "reopen",
		Usage: "reopen task",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagTasklistName, Aliases: []string{"l"}, Value: viper.GetString(txtDefaultList)},
			&cli.StringFlag{Name: flagGroupName, Value: viper.GetString(txtDefaultGroup)},
			&cli.StringFlag{Name: flagTitleName, Required: true},
			&cli.StringFlag{Name: flagUUIDName},
		},

		Action: func(c *cli.Context) error {
			if c.String(flagTasklistName) == "" && c.String(flagUUIDName) == "" {
				return fmt.Errorf("either --%s or --%s must be specified", flagTasklistName, flagUUIDName)
			}

			opts := getOpts(c)
			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			if c.String(flagGroupName) != "" {
				reopenTaskInput := sncli.ReopenAdvancedTaskInput{
					Session:  &sess,
					Tasklist: c.String(flagTasklistName),
					Group:    c.String(flagGroupName),
					Title:    c.String(flagTitleName),
					UUID:     c.String(flagUUIDName),
				}

				if err = reopenTaskInput.Run(); err != nil {
					return err
				}
			} else {
				reopenTaskInput := sncli.ReopenTaskInput{
					Session:  &sess,
					Tasklist: c.String(flagTasklistName),
					Title:    c.String(flagTitleName),
					UUID:     c.String(flagUUIDName),
				}

				if err = reopenTaskInput.Run(); err != nil {
					return err
				}
			}

			fmt.Println(msgTaskReopened)

			return nil
		},
	}
}
