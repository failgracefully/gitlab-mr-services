package main

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"
)

func HandleMerge(request requestBody, git *gitlab.Client) int {

	// set a new client
	var mr_labels []string
	var rec_labels []string
	var find_val, replace_var string

	// check if which kind of label we need to handle
	if request.Project.Id == 1907 {
		find_val = "R-"
		replace_var = "MM-"
	} else {
		find_val = "MM-"
		replace_var = "R-"
	}

	// split to two search groups
	for _, label := range request.Labels {
		if IsValidUUID(label.Title) {
			mr_labels = append(mr_labels, label.Title)
			rec_labels = append(rec_labels, strings.Replace(label.Title, find_val, replace_var, -1))
		}
	}

	printSlice(mr_labels)

	// query for merge requests with the first given search group labels
	opt := gitlab.ListMergeRequestsOptions{Scope: gitlab.String("all"), State: gitlab.String("opened"), Labels: mr_labels}
	mergerequests, resp, err := git.MergeRequests.ListMergeRequests(&opt, nil)

	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	// query for merge requests with the second given search group labels
	rec_opt := gitlab.ListMergeRequestsOptions{Scope: gitlab.String("all"), State: gitlab.String("opened"), Labels: rec_labels}
	rec_mergerequests, rec_resp, rec_err := git.MergeRequests.ListMergeRequests(&rec_opt, nil)

	if rec_err != nil {
		fmt.Println(rec_err.Error())
		return 1
	}

	// merge two sets of merge requests into one
	if rec_resp.Status == "200 OK" {
		for _, rec_mr := range rec_mergerequests {
			mergerequests = append(mergerequests, rec_mr)
		}
	}

	// found linked merge requests
	if len(mergerequests) > 0 && resp.Status == "200 OK" {
		fmt.Printf("Found %d Merge Requests \n", len(mergerequests))

		// Check that there are no merge requests from the same project
		var mr_projects []int
		for _, mr := range mergerequests {
			b, index := in_array(mr.ProjectID, mr_projects)
			if b {
				fmt.Printf("Found two linked merge requests in project %d, cancelling merge...", mr.ProjectID)
				return index
			} else {
				mr_projects = append(mr_projects, mr.ProjectID)
			}
		}

		// for each of the merge requests, make sure they can be automerged without any errors
		bCanBeMerged := true
		for _, mr := range mergerequests {
			fmt.Printf("Merge Request %d!(%d):\"%s\" merge status: %s \n", mr.IID, mr.ID, mr.Title, mr.MergeStatus)
			if mr.MergeStatus == "cannot_be_merged" {
				bCanBeMerged = false
				return 1
			}
		}

		//merge them
		if bCanBeMerged {
			for _, mr := range mergerequests {
				if mr.ID != request.ObjectAttributes.Id && mr.State != "merged" {
					amropt := gitlab.AcceptMergeRequestOptions{}
					_, _, err = git.MergeRequests.AcceptMergeRequest(mr.ProjectID, mr.IID, &amropt, nil)

					if err != nil {
						fmt.Printf("Failed merging Merge Request %d. Error: %s", mr.ID, err.Error())
						return 1
					}
				}
			}
		} else {
			fmt.Println("OH MAN I CANT MERGE")
		}

		// if no linked merge requests were found
	} else if len(mergerequests) == 0 {
		fmt.Printf("No Merge Requests with the given labels were found")
	}

	return 0
}
