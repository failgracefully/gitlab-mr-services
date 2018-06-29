package main

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

func HandleMerge(request requestBody, git *gitlab.Client) int {

	// set a new client
	var mr_labels []string

	for _, label := range request.Labels {
		mr_labels = append(mr_labels, label.Title)
	}

	printSlice(mr_labels)

	// query for merge requests with the given labels
	opt := gitlab.ListMergeRequestsOptions{Scope: gitlab.String("all"), State: gitlab.String("opened"), Labels: mr_labels}
	mergerequests, resp, err := git.MergeRequests.ListMergeRequests(&opt, nil)

	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	// found linked merge requests
	if len(mergerequests) > 0 && resp.Status == "200 OK" {
		fmt.Printf("Found %d Merge Requests \n", len(mergerequests))

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
