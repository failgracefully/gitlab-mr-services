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

	http.HandleFunc("/router", func(w http.ResponseWriter, r *http.Request) {
		queryPrivateToken := r.URL.Query().Get("private_token")
		var privateToken *string
		if queryPrivateToken != "" {
			privateToken = &queryPrivateToken
		} else {
			privateToken = privateTokenGlobal
		}
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
		//		log.Printf("INFO: Received %s", string(requestBodyAsByteArray))
		log.Printf("[ROUTER] INFO: Received MR: %d with action %s", requestBody.ObjectAttributes.Id, requestBody.ObjectAttributes.Action)
		callRest(requestBodyAsByteArray)

		git := gitlab.NewClient(nil, string([]byte(*privateToken)))
		git.SetBaseURL(*baseURL)

		// in case we opened a new merge request
		if requestBody.ObjectAttributes.Action == "open" {

			// label it with unique label
			log.Printf("[ROUTER] Handle labeling for MR %d", requestBody.ObjectAttributes.Id)
			HandleLabel(*requestBody, git)

			// call BOMR

			// in case we merging a mergerequest
		} else if requestBody.ObjectAttributes.Action == "merge" {

			log.Printf("[ROUTER] Handle merging for MR %d", requestBody.ObjectAttributes.Id)
			// merge linked mergerequests
			HandleMerge(*requestBody, git)

			// in case the merge request was updated
		} else if requestBody.ObjectAttributes.Action == "update" {
			//Originally called the pipeline, currently lives side by side
			//call bomr
		}
	})

	log.Printf(fmt.Sprintf("[ROUTER] INFO: Listening on port %d", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func callRest(jsonStr []byte) int {
	url := "http://localhost:8080/hook"
	log.Printf("[ROUTER] sending to bomr")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("[ROUTER] response Status:", resp.Status)
	log.Println("[ROUTER] response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("[ROUTER] response Body:", string(body))

	return 0
}
