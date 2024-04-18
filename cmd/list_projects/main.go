package main

import(
	"fmt"
	"log"

	"github.com/jalitriver/gitlab-cmds/cmd/internal"
	"github.com/xanzy/go-gitlab"
)

func main() {

	// FIXME: Should be option in config.json.
	baseURL := "https://gitlab.serice.net/api/v4"
	
	// FIXME: Should be option in config.json.
	fname := "/home/pserice/src/go/gitlab-cmds.git/auth.json"

	// Load the authentication information from file.
	authInfo, err := internal.LoadAuthInfo(fname)
	if err != nil {
		log.Fatalf(
			"LoadAuthInfo: Unable to load authentication information " +
			"from file %v: %v", fname, err)
	}

	// Create the Gitlab client based on the authentication
	// information provided by the user.
	client, err := authInfo.CreateGitlabClient(
		gitlab.WithBaseURL(baseURL))
	if err != nil {
		log.Fatalf("CreateGitlabClient: %v\n", err)
	}

	// Get the list of projects.
	opts := gitlab.ListProjectsOptions{}
	ps, _, err := client.Projects.ListProjects(&opts)
	if err != nil {
		log.Fatalf("ListProjects: %v\n", err)
	}

	// Print each project.
	for _, p := range ps {
		fmt.Printf("%v\n", p)
	}
}
