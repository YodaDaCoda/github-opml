package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type Repo struct {
	FullName string `json:"full_name"`
}

var output string

func main() {
	// reference: https://gist.github.com/Sanjo/4ed367c68acc27fd9a18

	// parse arguments (cobra)
	var rootCmd = &cobra.Command{
		Use: "github-opml",
	}
	var cmdStarred = &cobra.Command{
		Use:   "starred [username]",
		Short: "Generate OPML file for releases of starred repos of given user",
		Run:   processStarred,
	}
	cmdStarred.Flags().StringVarP(&output, "output", "o", "", "Output file [default: stdout]")
	rootCmd.AddCommand(cmdStarred)
	rootCmd.Execute()
}

func processStarred(cmd *cobra.Command, args []string) {
	// read json API (loop)
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "No username given. Call with --help for details on usage.\n")
		os.Exit(1)
	}
	username := args[0]
	page := 1
	resultsPerPage := 100
	starRepos := []Repo{}
	for true {
		url := fmt.Sprintf("https://api.github.com/users/%s/starred?page=%d&per_page=%d", username, page, resultsPerPage)

		res, err := http.Get(url)
		perror(err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		perror(err)

		var repos []Repo
		err = json.Unmarshal(body, &repos)
		for _, repo := range repos {
			starRepos = append(starRepos, repo)
		}

		if len(repos) < resultsPerPage {
			break
		}
		page++
	}
	// build OPML
	fmt.Fprintf(os.Stderr, "Number of starred repos: %d\n", len(starRepos))
	// TODO use xml package
	xmlString :=
fmt.Sprintf(`<?xml version="1.0" encoding="utf-8" ?>
<opml version="2.0">
	<head>
		<title>%s's Github Starred Projects Releases</title>
		<dateCreated>%s</dateCreated>
	</head>
	<body>
		<outline text="GitHub Starred Releases">
`, username, time.Now().Format(time.RFC3339))
	for _, repo := range starRepos {
		releaseUrl := fmt.Sprintf("https://github.com/%s/releases.atom", repo.FullName)
		xmlString += fmt.Sprintf("\t\t\t<outline type=\"rss\" title=\"%s\" text=\"%s\" xmlUrl=\"%s\" />\n", repo.FullName, repo.FullName, releaseUrl)
	}
	xmlString +=
`		</outline>
	</body>
</opml>`

	// output to stdout or file
	if output != "" {
		ioutil.WriteFile(output, []byte(xmlString), 0640)
	} else {
		fmt.Println(xmlString)
	}
}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}
