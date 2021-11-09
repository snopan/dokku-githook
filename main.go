package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const PLUGIN_NAME = "github-hook"

func readLocalHooksData() ([]string, error) {
	hooks := make([]string, 0)

	// Retrieve hooks from the local data storage
	// hookPath := fmt.Sprintf("%s/%s/data/hooks", os.Getenv("PLUGIN_AVAILABLE_PATH"), PLUGIN_NAME)
	hookFile, err := os.Open("./data/hooks")
	if err != nil {
		return hooks, fmt.Errorf("error loading hooks: %w", err)
	}

	// Loop through each line and retrieve the hook
	hookScanner := bufio.NewScanner(hookFile)
	for hookScanner.Scan() {

		// Each line is in the format "hook webhookId repositoryShort"
		hookLine := hookScanner.Text()
		hookArr := strings.Fields(hookLine)
		hook := hookArr[0]

		// Store the hook
		hooks = append(hooks, hook)
	}
	if hookScanner.Err() != nil {
		return hooks, fmt.Errorf("error parsing hooks: %w", err)
	}

	return hooks, nil
}

func readLocalLinksData() (map[string][]string, error) {
	links := make(map[string][]string)

	// Retrieve links from the local data storage
	// linkPath := fmt.Sprintf("%s/%s/data/links", os.Getenv("PLUGIN_AVAILABLE_PATH"), PLUGIN_NAME)
	linkFile, err := os.Open("./data/links")
	if err != nil {
		return links, fmt.Errorf("error loading links: %w", err)
	}

	// Loop through each line and retrieve the hook and app
	linkScanner := bufio.NewScanner(linkFile)
	for linkScanner.Scan() {

		// Each line is in the format "hook app"
		linkLine := linkScanner.Text()
		linkArr := strings.Fields(linkLine)
		hook := linkArr[0]
		app := linkArr[1]

		// When no apps are stored under a hook, initialize the an array
		if _, ok := links[hook]; !ok {
			links[hook] = make([]string, 0)
		}

		// Store hook as key and app in an array as value
		links[hook] = append(links[hook], app)
	}
	if linkScanner.Err() != nil {
		return links, fmt.Errorf("error parsing links: %w", err)
	}

	return links, nil
}

func readLocalDeploysData() (map[string]string, error) {
	deploys := make(map[string]string)

	// Retrieve deploys from the local data storage
	// deployPath := fmt.Sprintf("%s/%s/data/deploys", os.Getenv("PLUGIN_AVAILABLE_PATH"), PLUGIN_NAME)
	deployFile, err := os.Open("./data/deploys")
	if err != nil {
		return deploys, fmt.Errorf("error loading deploys: %w", err)
	}

	// Loop through each line and retrieve the app and repository
	deployScanner := bufio.NewScanner(deployFile)
	for deployScanner.Scan() {

		// Each line is in the format "app repository"
		deployLine := deployScanner.Text()
		deployArr := strings.Fields(deployLine)
		app := deployArr[0]
		repository := deployArr[1]

		// Store app as key and repository as value
		deploys[app] = repository
	}
	if deployScanner.Err() != nil {
		return deploys, fmt.Errorf("error parsing links: %w", err)
	}

	return deploys, nil
}

func deployApp(app string, repository string) {
	cmd := exec.Command("dokku","git:sync", "--build", app, repository)

	// Pipe the output of the command to stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("fatal error piping command output: %s", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatalf("fatal error running command: %s", err)
	}

	// Read the data from stdout
	scanner := bufio.NewScanner(stdout)

	// Print each line
	log.Print(fmt.Sprintf(`App deployment log for "%s"`, app))	
	for scanner.Scan() {
		log.Print(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("fatal error waiting for command to finish: %s", err)
	}

	log.Printf(`App "%s" has been deployed!`, app)
}

func main() {
	// Read all the local data
	hookArr, err := readLocalHooksData()
	if err != nil {
		log.Fatalf("fatal error reading local hook data: %s", err)
	}
	log.Print("Loaded local hooks data")

	linkDict, err := readLocalLinksData()
	if err != nil {
		log.Fatalf("fatal error reading local link data: %s", err)
	}
	log.Print("Loaded local links data")

	deployDict, err := readLocalDeploysData()
	if err != nil {
		log.Fatalf("fatal error reading local deploy data: %s", err)
	}
	log.Print("Loaded local deploys data")

	// For each hook, start listening for github requests
	for _, hook := range hookArr {
		http.HandleFunc(fmt.Sprintf("/%s", hook), func(w http.ResponseWriter, r *http.Request) {

			// When request comes in, find all the apps linked to the hook
			log.Print(fmt.Sprintf(`Hook "%s" was triggered`, hook))
			appArr := linkDict[hook]
			for _, app := range appArr {

				// Then deploy each app
				deployApp(app, deployDict[app])
			}
		})
	}

	// Start the http server
	log.Print(fmt.Sprintf("Starting the http server on port %s!", os.Getenv("GITHUB_HOOK_PORT")))
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("GITHUB_HOOK_PORT")), nil)
}
