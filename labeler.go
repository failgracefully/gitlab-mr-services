package main

import (
	"log"
	"os/exec"

	gitlab "github.com/xanzy/go-gitlab"
)

//"net/url"
//"strings"

func HandleLabel(request requestBody, git *gitlab.Client) int {

	// invent a new label (find somekind of formula)
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	uuid := string(out[:])
	// add this label to the project
	opt := gitlab.CreateLabelOptions{Name: gitlab.String(uuid), Color: gitlab.String("#0033CC")}
	created_label, resp, err := git.Labels.CreateLabel(request.Project.Id, &opt, nil)

	if err != nil {
		log.Fatal(err)
	}

	// add this label to the merge request
	//	var labels gitlab.Labels
	//	labels = append(labels, label)
	if resp.Status == "201 Created" {

		var mr_labels []string
		for _, label := range request.Labels {
			mr_labels = append(mr_labels, label.Title)
		}
		mr_labels = append(mr_labels, created_label.Name)

		printSlice(mr_labels)
		mruopt := gitlab.UpdateMergeRequestOptions{Labels: mr_labels}
		git.MergeRequests.UpdateMergeRequest(request.Project.Id, request.ObjectAttributes.Id, &mruopt, nil)

		// optional: add to the description a line that describe this label
	}
	return 0
}
