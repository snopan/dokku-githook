package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const PLUGIN_NAME = "github-hook"

type LocalData struct {
	hooks   []string
	links   map[string][]string
	deploys map[string]string

	mu sync.Mutex
}

var localData = &LocalData{}

func readLocalDataLines(filename string) ([]string, error) {
	var output []string

	// Load the file
	file, err := os.Open(fmt.Sprintf("./data/%s", filename))
	if err != nil {
		return output, fmt.Errorf("error opening file: %w", err)
	}

	// Read line by line and append it to a string slice
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		output = append(output, fileScanner.Text())
	}

	// Check if the file scanner has failed
	if err := fileScanner.Err(); err != nil {
		return output, fmt.Errorf("error scanning file: %w", err)
	}

	return output, nil
}

func readLocalHooksData() ([]string, error) {
	var hooks []string

	// Read the hooks file
	hookLines, err := readLocalDataLines("hooks")
	if err != nil {
		return hooks, fmt.Errorf("error loading hooks file: %w", err)
	}

	// Parse each line and store the data
	for _, hl := range hookLines {
		hookArr := strings.Fields(hl)
		hook := hookArr[0]
		hooks = append(hooks, hook)
	}

	return hooks, nil
}

func readLocalLinksData() (map[string][]string, error) {
	var links map[string][]string

	// Read the links file
	linkLines, err := readLocalDataLines("links")
	if err != nil {
		return links, fmt.Errorf("error loading links file: %w", err)
	}

	// Parse each line and store the data
	for _, ll := range linkLines {
		linkArr := strings.Fields(ll)
		hook := linkArr[0]
		app := linkArr[1]

		// When no apps are stored under a hook, initialize the an array
		if _, ok := links[hook]; !ok {
			links[hook] = make([]string, 0)
		}

		// Store hook as key and app in an array as value
		links[hook] = append(links[hook], app)
	}

	return links, nil
}

func readLocalDeploysData() (map[string]string, error) {
	var deploys map[string]string

	// Read the links file
	deployLines, err := readLocalDataLines("deploys")
	if err != nil {
		return deploys, fmt.Errorf("error loading deploys file: %w", err)
	}

	// Parse each line and store the data
	for _, dl := range deployLines {
		deployArr := strings.Fields(dl)
		app := deployArr[0]
		repository := deployArr[1]
		deploys[app] = repository
	}

	return deploys, nil
}

func deployApp(app string, repository string) error {
	cmd := exec.Command("dokku", "git:sync", "--build", app, repository)

	// Write the stdout and stderr output of the command to separate buffers
	var stdoutBuff, stderrBuff bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuff)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuff)

	// Start the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	// Read the data from stdout and stderr
	outStr, errStr := stdoutBuff.String(), stderrBuff.String()
	log.Printf("out: %s\nerr: %s\n", outStr, errStr)

	// Check if stderr exists
	if len(errStr) == 0 {
		log.Printf("App %s has been deployed", app)
		return nil
	} else {
		return fmt.Errorf("error executing dokku deploy: %s", errStr)
	}
}

func (ld *LocalData) loadAll() error {
	defer ld.mu.Unlock()
	ld.mu.Lock()

	// Read all the local data
	hookArr, err := readLocalHooksData()
	if err != nil {
		return fmt.Errorf("error reading local hooks data: %w", err)
	}
	linkDict, err := readLocalLinksData()
	if err != nil {
		return fmt.Errorf("error reading local links data: %w", err)
	}

	deployDict, err := readLocalDeploysData()
	if err != nil {
		return fmt.Errorf("error reading local deploys data: %w", err)
	}

	// Store the local data read
	ld.hooks = hookArr
	ld.links = linkDict
	ld.deploys = deployDict
	return nil
}

func (ld *LocalData) deployAll() error {
	for app := range ld.deploys {
		if err := deployApp(app, ld.deploys[app]); err != nil {
			return fmt.Errorf("error deploying app %s: %w", app, err)
		}
	}
	return nil
}

func runHookServer(ld *LocalData) {
	var hookServer *http.ServeMux

	hookServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hook := r.URL.Path[1:]

		// Check if the request is for a hook
		if appArr, ok := ld.links[hook]; ok {
			log.Printf(`Hook "%s" was triggered`, hook)

			// Then deploy each app
			for _, app := range appArr {
				if err := deployApp(app, ld.deploys[app]); err != nil {
					log.Printf("error deploying app %s: %s", app, err)
				}
			}
		}
	})

	log.Printf("Starting hook server on port %s", os.Getenv("GITHUB_HOOK_PORT"))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("GITHUB_HOOK_PORT")), hookServer); err != nil {
		log.Fatalf("error to starting control server: %s", err)
	}
}

func runControlServer(ld *LocalData) {
	var controlServer *http.ServeMux

	controlServer.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		// Load all the local data
		if err := localData.loadAll(); err != nil {
			log.Fatalf("error loading all local data: %s", err)
		}
		log.Print("Finished loading all local data")
	})

	log.Printf("Starting control server on port %s", os.Getenv("LOCAL_CONTROL_PORT"))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("LOCAL_CONTROL_PORT")), controlServer); err != nil {
		log.Fatalf("error to starting control server: %s", err)
	}
}

func main() {
	// Load all the local data
	if err := localData.loadAll(); err != nil {
		log.Fatalf("error loading all local data: %s", err)
	}
	log.Print("Finished loading all local data")

	// Make an inital deploy for all the apps that have deployment set
	if err := localData.deployAll(); err != nil {
		log.Fatalf("error deploying all apps: %s", err)
	}
	log.Print("Finished deploying all apps")

	// Start hook server
	go runHookServer(localData)

	// Start control server
	runControlServer(localData)
}
