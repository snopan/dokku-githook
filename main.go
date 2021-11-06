package main

import (
	"net/http"
	"bufio"
	"os"
	"os/exec"
	"fmt"
	"log"
	"strings"
)

var PLUGIN_NAME="github-hook"

func checkErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func readLocalHooksData() []string {
	returnArr := make([]string, 0)

	// Retrieve hooks from the local data storage
	// hookPath := fmt.Sprintf("%s/%s/data/hooks", os.Getenv("PLUGIN_AVAILABLE_PATH"), PLUGIN_NAME)
	hookFile, err := os.Open("./data/hooks")
	checkErr(err)

	// Loop through each line and retrieve the hook
	hookScanner := bufio.NewScanner(hookFile)
	for hookScanner.Scan() {

		// Each line is in the format "hook webhookId repositoryShort"
		hookLine := hookScanner.Text()
		hookArr := strings.Fields(hookLine)
		hook := hookArr[0]

		// Store the hook
		returnArr = append(returnArr, hook)
	}
	checkErr(hookScanner.Err())
	return returnArr
}


func readLocalLinksData() map[string][]string {
	returnDict := make(map[string][]string)

	// Retrieve links from the local data storage
	// linkPath := fmt.Sprintf("%s/%s/data/links", os.Getenv("PLUGIN_AVAILABLE_PATH"), PLUGIN_NAME)
	linkFile, err := os.Open("./data/links")
	checkErr(err)

	// Loop through each line and retrieve the hook and app
	linkScanner := bufio.NewScanner(linkFile)
	for linkScanner.Scan() {

		// Each line is in the format "hook app"
		linkLine := linkScanner.Text()
		linkArr := strings.Fields(linkLine)
		hook := linkArr[0]
		app := linkArr[1]

		// When no apps are stored under a hook, initialize the an array
		if _, ok := returnDict[hook]; !ok {
			returnDict[hook] = make([]string, 0)
		}

		// Store hook as key and app in an array as value
		returnDict[hook] = append(returnDict[hook], app)
	}
	checkErr(linkScanner.Err())
	return returnDict
}

func readLocalDeploysData() map[string]string {
	returnDict := make(map[string]string)

	// Retrieve deploys from the local data storage
	// deployPath := fmt.Sprintf("%s/%s/data/deploys", os.Getenv("PLUGIN_AVAILABLE_PATH"), PLUGIN_NAME)
	deployFile, err := os.Open("./data/deploys")
	checkErr(err)

	// Loop through each line and retrieve the app and repository
	deployScanner := bufio.NewScanner(deployFile)
	for deployScanner.Scan() {

		// Each line is in the format "app repository"
		deployLine := deployScanner.Text()
		deployArr := strings.Fields(deployLine)
		app := deployArr[0]
		repository := deployArr[1]

		// Store app as key and repository as value
		returnDict[app] = repository
	}
	checkErr(deployScanner.Err())
	return returnDict
}

func main() {

	// Read all the local data
	hookArr := readLocalHooksData()
	log.Print("Loaded local hooks data")

	linkDict := readLocalLinksData()
	log.Print("Loaded local links data")

	deployDict := readLocalDeploysData()
	log.Print("Loaded local deploys data")

	// For each hook, start listening for github requests
	for _, hook := range hookArr {
		http.HandleFunc("/"+hook, func(w http.ResponseWriter, r *http.Request) {

			// When reuest comes, find all the apps linked to the hook
			log.Print(fmt.Sprintf("Hook \"%s\" was triggered", hook))
			appArr := linkDict[hook]
			for _, app :=range appArr {

				// Then deploy each app
				log.Print(fmt.Sprintf("App \"%s\" is being deployed", app))
				cmd := exec.Command("dokku", "--build", "git:sync", app, deployDict[app])
				cmd.Run()
				log.Print(fmt.Sprintf("App \"%s\" is deployed!", app))
			}
		})
	}

	// Start the http server
	log.Print(fmt.Sprintf("Trying to bind to port %s!", os.Getenv("GITHUB_HOOK_PORT")))
	http.ListenAndServe(":4000", nil)
	log.Print(fmt.Sprintf("Github hook http server has started on port %s!", os.Getenv("GITHUB_HOOK_PORT")))
}