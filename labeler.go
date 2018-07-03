package main

import (
	"fmt"
	"os/exec"

	gitlab "github.com/xanzy/go-gitlab"
)

func HandleLabel(request requestBody, git *gitlab.Client) int {

	// Create new guid to be used as label
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}
	uuid := string(out[:])

	// Add this label to the project
	opt := gitlab.CreateLabelOptions{Name: gitlab.String(uuid), Color: gitlab.String("#0033CC")}
	created_label, resp, err := git.Labels.CreateLabel(request.Project.Id, &opt, nil)

	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}

	// Label created successfully on server side
	if resp.Status == "201 Created" {

		var mr_labels []string

		// Append label to existing labels
		for _, label := range request.Labels {
			mr_labels = append(mr_labels, label.Title)
		}
		mr_labels = append(mr_labels, created_label.Name)

		// Print labels for debugging
		printSlice(mr_labels)

		// Update merge request with current labels with additional new label
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

// cleanlabels delete unused labels in given project
// Unused label is a label that has no open issues or merge requests
func cleanLabels(projectId int, git *gitlab.Client) int {

	// Query project's labels
	opt := gitlab.ListLabelsOptions{}
	projectLabels, _, err := git.Labels.ListLabels(projectId, &opt, nil)

	// Check for error
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}

	// For each label in the project
	for _, label := range projectLabels {

		// Check if label is unused
		if label.OpenIssuesCount > 0 || label.OpenMergeRequestsCount > 0 || label.ClosedIssuesCount > 0 {

			// Delete label is determined as unused
			dopt := gitlab.DeleteLabelOptions{Name: gitlab.String(label.Name)}
			_, err = git.Labels.DeleteLabel(projectId, &dopt, nil)

			// Check for deletion errors
			if err != nil {
				fmt.Printf(err.Error())
				return 1
			}
		}
	}

	return 0
}
