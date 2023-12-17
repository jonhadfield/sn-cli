package main

// {
// 	Name:   "export",
// 	Usage:  "export data",
// 	Hidden: true,
// 	Flags: []cli.Flag{
// 		cli.StringFlag{
// 			Name:  "path",
// 			Usage: "choose directory to place export in (default: current directory)",
// 		},
// 	},
// 	Action: func(c *cli.Context) error {
// 		var opts configOptsOutput
// 		opts, err = getOpts(c)
// 		if err != nil {
// 			return err
// 		}
// 		// useStdOut = opts.useStdOut
//
// 		outputPath := strings.TrimSpace(c.String("output"))
// 		if outputPath == "" {
// 			outputPath, err = os.Getwd()
// 			if err != nil {
// 				return err
// 			}
// 		}
//
// 		timeStamp := time.Now().UTC().Format("20060102150405")
// 		filePath := fmt.Sprintf("standard_notes_export_%s.json", timeStamp)
// 		outputPath += string(os.PathSeparator) + filePath
//
// 		var sess cache.Session
// 		sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
// 		if err != nil {
// 			return err
// 		}
//
// 		var cacheDBPath string
// 		cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
// 		if err != nil {
// 			return err
// 		}
//
// 		sess.Debug = opts.debug
//
// 		sess.CacheDBPath = cacheDBPath
// 		appExportConfig := sncli.ExportConfig{
// 			Session:   &sess,
// 			Decrypted: c.Bool("decrypted"),
// 			File:      outputPath,
// 		}
// 		err = appExportConfig.Run()
// 		if err == nil {
// 			msg = fmt.Sprintf("encrypted export written to: %s", outputPath)
// 		}
//
// 		return err
// 	},
// },
// {
// 	Name:   "import",
// 	Usage:  "import data",
// 	Hidden: true,
// 	Flags: []cli.Flag{
// 		cli.StringFlag{
// 			Name:  "file",
// 			Usage: "path of file to import",
// 		},
// 		cli.BoolFlag{
// 			Name:  "experiment",
// 			Usage: "test import functionality - only use after taking backup as this is experimental",
// 		},
// 	},
// 	Action: func(c *cli.Context) error {
// 		var opts configOptsOutput
// 		opts, err = getOpts(c)
// 		if err != nil {
// 			return err
// 		}
//
// 		// useStdOut = opts.useStdOut
//
// 		inputPath := strings.TrimSpace(c.String("file"))
// 		if inputPath == "" {
// 			return errors.New("please specify path using --file")
// 		}
//
// 		if !c.Bool("experiment") {
// 			fmt.Printf("\nWARNING: The import functionality is currently for testing only\nDo not use unless you have a backup of your data and intend to restore after testing\nTo proceed run the command with flag --experiment\n")
// 			return nil
// 		}
//
// 		var session cache.Session
// 		session, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
// 		if err != nil {
// 			return err
// 		}
//
// 		session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
// 		if err != nil {
// 			return err
// 		}
//
// 		appImportConfig := sncli.ImportConfig{
// 			Session:   &session,
// 			File:      inputPath,
// 			Format:    c.String("format"),
// 			Debug:     opts.debug,
// 			UseStdOut: opts.useStdOut,
// 		}
//
// 		var imported int
// 		imported, err = appImportConfig.Run()
// 		if err == nil {
// 			msg = fmt.Sprintf("imported %d items", imported)
// 		} else {
// 			msg = "import failed"
// 		}
//
// 		return err
// 	},
// },
