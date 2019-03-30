package main

import (
	"errors"
	"fmt"
	"github.com/rezen/query"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the info severity or above.
	log.SetLevel(log.ErrorLevel)
}

func main() {
	searcher := query.DefaultQueryer()
	var txn query.Transaction
	var target string
	var querystring string
	root := &cobra.Command{
		Use:   "awwwditq",
		Short: "A brief description of your application",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(target) < 4 {
				return errors.New("Invalid target")
			}

			if len(querystring) < 3 {
				return errors.New("Need a valid query")
			}

			t := query.TargetFromString(target)
			q := query.StringToQuery(querystring)
			q.Target = t

			err := searcher.Validate(q)

			if err != nil {
				return err
			}

			txn = searcher.Exec(q)

			if txn.Error != nil {
				return txn.Error
			}

			fmt.Println("Results ...")
			fmt.Println("------------------------------------")

			for _, result := range txn.GetResults() {
				switch result.(type) {
				default:
					fmt.Println(result.AsText())
				case query.TextResult:
					fmt.Println(result.AsText())
				}
			}
			return nil
		},
	}

	/*
		// ??
		root.SetFlagErrorFunc(func(*cobra.Command, error) error {
			fmt.Println("ERRRRRRR")
			return nil
		})
	*/

	root.PersistentFlags().StringVarP(&target, "target", "t", "http://example.com", "Target for executing queries against")
	root.PersistentFlags().StringVarP(&querystring, "query", "q", "", "Query for info (-q \"http > doc > title\")")
	if err := root.Execute(); err != nil {
		last := len(txn.Executed) - 1

		fmt.Println(err)
		if err == query.ErrorMissingResource || err == query.ErrorNoExecutor {
			fmt.Println("Try one of these:", searcher.Selectable())
		} else {

			if last >= 0 {
				fmt.Println("Try using one of these:", txn.Executed[last].Selectable())
			} else {
				fmt.Println("Try one of these:", searcher.Selectable())
			}
		}

		os.Exit(1)
	}

	/*
		pagesMap := map[string][]string{
			"*": {
				"/",
				"robots.txt",
				"sitemap.xml",
				".gitignore",
				".git/HEAD",
				".env",
				"package.json",
				"crossdomain.xml",
			},
			"php": {
				"composer.json",
				"phpinfo.php",
				"php.ini",
				"phpmyadmin/",
			},
			"python": {
				"requirements.txt",
				"Pipfile",
				"setup.py",
			},
			"java": {
				"pom.xml",
				"web-console/",
			},
			"nodejs": {
				".npmrc",
			},
			"ruby": {
				"Gemfile",
				"Gemfile.lock",
				"Rakefile",
				"app/controllers/application.rb",
				"config.ru",
			},
			"extras": {
				".travis.yml",
				"error_log",
				"vendor/",
				"api/",
				"config/",
				"backups/",
			},
		}*/

}
