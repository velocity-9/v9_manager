package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func init() {
	// Setup log streams
	setLogStreams(os.Stdout, os.Stdout, os.Stderr)
}

func main() {
	//Initialize default ports
	CIPort := "0.0.0.0:81"
	websitePort := "0.0.0.0:80"

	// Check for development flag
	if len(os.Args) > 1 && contains(os.Args, "--development") {
		CIPort = ":3081"
		websitePort = ":3080"
	}

	// Get workers from env
	workers, envErr := getWorkers()
	if envErr != nil {
		Error.Println("Error getting worker info", envErr)
		return
	}

	// Get psql info from env
	psqlInfo, psqlInfoErr := getPsqlInfo()
	if psqlInfoErr != nil {
		Error.Println("Error getting psql info", psqlInfoErr)
		return
	}

	Info.Println("CIPort", CIPort, "websitePort", websitePort, "workers", workers)

	dbErr := SetupDatabasePopulator(psqlInfo, workers)
	if dbErr != nil {
		Error.Println("Error connecting to DB", dbErr)
		return
	}

	http.Handle("/payload", &pushHandler{workers: workers, counter: 0, deployer: &Deployer{
		allWorkers:             workers,
		deploymentChannelMutex: sync.RWMutex{},
		deploymentChannels: make(map[repoPath]chan deploymentInfo),
	}})
	Info.Println("Starting Server...")
	err := http.ListenAndServe(CIPort, nil)
	if err != nil {
		Error.Println("CI http.ListenAndServe Error:", err)
	}
}

// Get env variables
func getEnvVar(name string) (string, error) {
	val, exists := os.LookupEnv(name)
	if !exists {
		return "", errors.New("Missing env variable: " + name)
	}

	return val, nil
}

func getWorkers() ([]*V9Worker, error) {
	workerString, err := getEnvVar("V9_WORKERS")
	if err != nil {
		return nil, err
	}

	workerUrls := strings.Split(workerString, ";")
	var workers = make([]*V9Worker, len(workerUrls))

	for i, url := range workerUrls {
		workers[i] = &V9Worker{url: url}
	}
	return workers, nil
}

func getPsqlInfo() (string, error) {
	pgHost, err := getEnvVar("V9_PG_HOST")
	if err != nil {
		return "", err
	}

	pgPortString, err := getEnvVar("V9_PG_PORT")
	if err != nil {
		return "", err
	}
	pgPort, err := strconv.Atoi(pgPortString)
	if err != nil {
		return "", fmt.Errorf("err: V9_PG_PORT must be a valid integer, was %s: %w", pgPortString, err)
	}

	pgUser, err := getEnvVar("V9_PG_USER")
	if err != nil {
		return "", err
	}

	pgPassword, err := getEnvVar("V9_PG_PASSWORD")
	if err != nil {
		return "", err
	}

	pgDb, err := getEnvVar("V9_PG_DB")
	if err != nil {
		return "", err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPassword, pgDb)

	return psqlInfo, nil
}

// FIXME: this should be in the helper class
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
