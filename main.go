package main

import (
	"net/http"
	"os"
)

func init() {
	// Setup log streams
	setLogStreams(os.Stdout, os.Stdout, os.Stderr)

}

func main() {
	// Get Environment variables
	CIPort, websitePort, worker := getEnvVariables()
	if CIPort == "" || websitePort == "" || worker == nil {
		Error.Println("Error loading env variables")
		return
	}

	go func() {
		Info.Println("Starting status handler...")
		http.Handle("/status", &statusHandler{worker: worker})
		err := http.ListenAndServe(websitePort, nil)
		if err != nil {
			Error.Println("Status http.ListenAndServer Error:", err)
		}
	}()

	http.Handle("/payload", &pushHandler{worker: worker, counter: 0})
	Info.Println("Starting Server...")
	err := http.ListenAndServe(CIPort, nil)
	if err != nil {
		Error.Println("CI http.ListenAndServe Error:", err)
	}
}

// Get env variables
func getEnvVariables() (string, string, []string) {
	CIPort, exists := os.LookupEnv("CI_PORT")
	if !exists {
		Error.Println("Failed to find CI_PORT")
		return "", "", nil
	}
	websitePort, exists := os.LookupEnv("WEBSITE_PORT")
	if !exists {
		Error.Println("Failed to find WEBSITE_PORT")
		return "", "", nil
	}

	workerArr := make([]string, 2, 5)
	worker, exists := os.LookupEnv("WORKER1")
	if !exists {
		Error.Println("Failed to find Worker URL")
		return "", "", nil
	}
	workerArr[0] = worker

	worker, exists = os.LookupEnv("WORKER2")
	if !exists {
		Error.Println("Failed to find Worker URL")
		return "", "", nil
	}
	workerArr[1] = worker

	return CIPort, websitePort, workerArr
}
