package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

type requestBody struct {
	ObjectKind string `json:"object_kind"` // merge_request
	Project    struct {
		Name string `json:"name"`
		Id   int    `json:"id"`
	} `json:"project"`
	ObjectAttributes struct {
		SourceBranch    string `json:"source_branch"`
		SourceProjectId int    `json:"source_project_id"`
		Id              int    `json:"id"`
		Iid             int    `json:"iid"`
		State           string `json:"state"` // merged, opened or closed
		Action          string `json:"action"`
		LastCommit      struct {
			Id string `json:"id"`
		} `json:"last_commit"`
		WorkInProgress bool `json:"work_in_progress"`
	} `json:"object_attributes"`
	Labels []label `json:"labels"`
}

type label struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
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

func printSlice(s []string) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}
