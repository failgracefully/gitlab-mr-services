package main

import (
	"fmt"
	"os/exec"

	gitlab "github.com/xanzy/go-gitlab"
)

func HandleLabel(request requestBody, git *gitlab.Client) int {

	// invent a new label (find somekind of formula)
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}
	uuid := string(out[:])
	// add this label to the project
	opt := gitlab.CreateLabelOptions{Name: gitlab.String(uuid), Color: gitlab.String("#0033CC")}
	created_label, resp, err := git.Labels.CreateLabel(request.Project.Id, &opt, nil)

	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}

	if resp.Status == "201 Created" {

		var mr_labels []string
		for _, label := range request.Labels {
			mr_labels = append(mr_labels, label.Title)
		}
		mr_labels = append(mr_labels, created_label.Name)

		printSlice(mr_labels)
		mruopt := gitlab.UpdateMergeRequestOptions{Labels: mr_labels}
		fmt.Printf("Projectid: %d, mergerequestid: %d, %+v \n", request.Project.Id, request.ObjectAttributes.Id, &mruopt)
		_, resp, err := git.MergeRequests.UpdateMergeRequest(request.Project.Id, request.ObjectAttributes.Iid, &mruopt, nil)

		if err != nil {
			fmt.Printf(err.Error())
			return 1
		}
		fmt.Printf(resp.Status)
	}
	return 0
}
