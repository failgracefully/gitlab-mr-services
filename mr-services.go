package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	gitlab "github.com/xanzy/go-gitlab"
)

func main() {
	var baseURL = flag.String("url", "", "URL (e.g. http://gitlab.com)")
	var privateTokenGlobal = flag.String("private_token", "", "Authorization Token (e.g. XXxXXx0xxxXXXxXxXxxX)")
	var port = flag.Int("port", 8181, "Port")

	flag.Parse()
	if *baseURL == "" {
		printUsageAndExit("Error: --url is required")
	}
	if *privateTokenGlobal == "" {
		printWarning("Warning: --private_token is not set")
	}

	http.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
		queryPrivateToken := r.URL.Query().Get("private_token")
		var privateToken *string

		// Override token given at service startup if defined in WebHook
		if queryPrivateToken != "" {
			privateToken = &queryPrivateToken
		} else {
			privateToken = privateTokenGlobal
		}

		// In case token wasnt given at WebHook or at service init
		if *privateToken == "" {
			fmt.Fprintf(os.Stderr, "Error: private_token is required\n")
		}
		var requestBody = &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Printf("WARN: Failed to deserialize request body (%s)", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		requestBodyAsByteArray, _ := json.Marshal(requestBody)
		log.Printf("[ROUTER] INFO: Received MR: %d with action: %s", requestBody.ObjectAttributes.Id, requestBody.ObjectAttributes.Action)

		git := gitlab.NewClient(nil, string([]byte(*privateToken)))
		git.SetBaseURL(*baseURL)

		// Determine what to do based on request's action
		if requestBody.ObjectAttributes.Action == "open" {

			// Label it with unique label
			log.Printf("[ROUTER] Handle labeling for MR %d", requestBody.ObjectAttributes.Id)
			HandleLabel(*requestBody, git)

			// Call ci build service
			callBomr(requestBodyAsByteArray)

		} else if requestBody.ObjectAttributes.Action == "merge" {

			log.Printf("[ROUTER] Handle merging for MR %d", requestBody.ObjectAttributes.Id)
			// Merge linked mergerequests
			HandleMerge(*requestBody, git)

		} else if requestBody.ObjectAttributes.Action == "update" {

			// Call ci build service
			callBomr(requestBodyAsByteArray)
		}
	})

	log.Printf(fmt.Sprintf("[ROUTER] INFO: Listening on port %d", *port))

	// Open channel for incoming requests with specific port
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

// Redirect merge requests to bomr service
func callBomr(jsonStr []byte) int {

	// address of bomr service inside docker container
	url := "http://localhost:8080/hook"
	log.Printf("[ROUTER] sending to bomr")

	// create the POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// Open http channel and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// Close reader when fuction returns
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("[ROUTER] BOMR resp Status:", resp.Status)
	log.Println("[ROUTER] response Body:", string(body))

	return 0
}
