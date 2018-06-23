package services

import (
	"flag"
	"fmt"
	//"net/url"
	"os"
	//"strings"
)

type requestBody struct {
	ObjectKind string `json:"object_kind"` // merge_request
	Project    struct {
		Name string `json:"name"`
	} `json:"project"`
	ObjectAttributes struct {
		SourceBranch    string `json:"source_branch"`
		SourceProjectId int    `json:"source_project_id"`
		Id              int    `json:"id"`
		State           string `json:"state"` // merged, opened or closed
		LastCommit      struct {
			Id string `json:"id"`
		} `json:"last_commit"`
		WorkInProgress bool `json:"work_in_progress"`
	} `json:"object_attributes"`
}

func printUsageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg+"\n\n")
	}
	flag.Usage()
	os.Exit(1)
}

func printWarning(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg+"\n\n")
	}
}
