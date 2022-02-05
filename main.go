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

const PLUGIN_PATH = "/var/lib/dokku/plugins/enabled/github-hook"

type LocalData struct {
	hooks   []string
	links   map[string][]string
	deploys map[string]string

	mu sync.Mutex
}

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
	links := make(map[string][]string)

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

		// Store hook as key and app in an array as value
		links[hook] = append(links[hook], app)
	}

	return links, nil
}

func readLocalDeploysData() (map[string]string, error) {
	deploys := make(map[string]string)

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

func deployApp(app string, repository string) error {
	logText(fmt.Sprintf("Deploying repostitory '%s' to app '%s'", repository, app))
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
	logCode(outStr)

	// Check if stderr exists
	if len(errStr) == 0 {
		log.Printf("App %s has been deployed", app)
		return nil
	} else {
		logText("error failed to deploy")
		logCode(errStr)
		return fmt.Errorf("error executing dokku deploy: %s", errStr)
	}
}

func logText(message string) {
	url := os.Getenv("DISCORD_WEBHOOK_URL")
	if len(url) == 0 {
		return
	}
	if _, err := exec.Command(fmt.Sprintf("bash -c source %s/logger.sh ; log %s %s", PLUGIN_PATH, url, message)).Output(); err != nil {
		log.Printf("error with text logger logging: %s: %s", message, err)
	}
}

func logCode(message string) {
	url := os.Getenv("DISCORD_WEBHOOK_URL")
	if len(url) == 0 {
		return
	}
	if _, err := exec.Command(fmt.Sprintf("bash -c source %s/logger.sh ; echo -n %s | logCode %s", PLUGIN_PATH, message, url)).Output(); err != nil {
		log.Printf("error with code logger logging: %s: %s", message, err)
	}
}

func runHookServer(ld *LocalData) {
	var hookServer http.ServeMux

	hookServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hook := r.URL.Path[1:]

		// Check if the request is for a hook
		if appArr, ok := ld.links[hook]; ok {
			log.Printf(`Hook "%s" was triggered`, hook)

			// Then deploy each app that is linked to the hook
			for _, app := range appArr {
				if err := deployApp(app, ld.deploys[app]); err != nil {
					log.Printf("error deploying app %s: %s", app, err)
				}
			}
		}
	})

	log.Printf("Starting hook server on port %s", os.Getenv("GITHUB_HOOK_PORT"))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("GITHUB_HOOK_PORT")), &hookServer); err != nil {
		log.Fatalf("error to starting control server: %s", err)
	}
}

func runControlServer(ld *LocalData) {
	log.Print("Reloading all local data")
	var controlServer http.ServeMux

	controlServer.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		// Load all the local data
		go func() {
			if err := ld.loadAll(); err != nil {
				log.Fatalf("error loading all local data: %s", err)
			} else {
				log.Print("Finished loading all local data")
			}
		}()
		w.WriteHeader(200)
	})

	controlServer.HandleFunc("/deploy-all", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Deploying all apps")

		// Deploy all apps with set repository
		go func() {
			for app, repository := range ld.deploys {
				log.Printf("Deploying repostitory '%s' to app '%s'", repository, app)
				if err := deployApp(app, repository); err != nil {
					log.Printf("error deploying app %s: %s", app, err)
				}
			}
		}()
		w.WriteHeader(200)
	})

	log.Printf("Starting control server on port %s", os.Getenv("LOCAL_CONTROL_PORT"))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("LOCAL_CONTROL_PORT")), &controlServer); err != nil {
		log.Fatalf("error to starting control server: %s", err)
	}
}

func main() {
	localData := &LocalData{}

	// Load all the local data
	if err := localData.loadAll(); err != nil {
		log.Fatalf("error loading all local data: %s", err)
	}
	log.Print("Finished loading all local data")

	// Start hook server
	go runHookServer(localData)

	// Start control server
	runControlServer(localData)
}
