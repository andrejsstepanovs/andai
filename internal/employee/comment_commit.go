package employee

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/andrejsstepanovs/andai/internal/exec"
)

func (i *Routine) commentCommitsSince(currentCommitSku, commitMessage string) error {
	newCommits, getShaErr := i.workbench.GetCommitsSinceInReverseOrder(currentCommitSku)

	//commits, getShaErr := i.workbench.GetBranchCommits(10)
	if getShaErr != nil {
		log.Printf("Failed to get last commit sha: %v", getShaErr)
		return nil
	}

	//if len(commits) == 0 {
	//	log.Println("No commits found")
	//	return nil
	//}
	//
	//newCommits := make([]string, 0)
	//old := true
	//for _, sha := range commits {
	//	if sha == currentCommitSku {
	//		old = false
	//		continue
	//	}
	//	if old {
	//		continue
	//	}
	//	newCommits = append(newCommits, sha)
	//}

	if len(newCommits) == 0 {
		log.Println("No new commits found")
		return nil
	}

	branchName := i.workbench.GetIssueBranchName(i.issue)
	format := "### Branch [%s](/projects/%s/repository/%s?rev=%s)"
	txt := make([]string, 0)
	txt = append(txt, fmt.Sprintf(format, branchName, i.project.Identifier, i.project.Identifier, branchName))
	for n, sha := range newCommits {
		format = "%d. Commit [%s](/projects/%s/repository/%s/revisions/%s/diff) - %s"
		txt = append(txt, fmt.Sprintf(format, n+1, sha, i.project.Identifier, i.project.Identifier, sha, commitMessage))
	}

	err := i.AddComment(strings.Join(txt, "\n"))
	if err != nil {
		return err
	}

	return nil
}

func (i *Routine) commitUncommitted(commitMessage string) (exec.Output, error) {
	modified := "git status | cat | grep modified | awk '{print $2}'"
	out, err := exec.Exec(modified, time.Minute)
	if err != nil {
		return exec.Output{}, err
	}
	if out.Stdout == "" {
		log.Println("No files to add")
		return exec.Output{}, nil
	}
	files := strings.Split(out.Stdout, "\n")
	if len(files) == 0 {
		log.Println("No files to add")
		return exec.Output{}, nil
	}

	lastCommit, err := i.workbench.GetLastCommit()
	if err != nil {
		return exec.Output{}, err
	}

	for _, f := range files {
		ret, err := exec.Exec(fmt.Sprintf("git add %s", f), time.Minute)
		if err != nil {
			return ret, err
		}
	}
	ret, err := exec.Exec("git commit -m \"code reformat\"", time.Minute)
	if err != nil {
		return ret, err
	}
	err = i.commentCommitsSince(lastCommit, commitMessage)
	if err != nil {
		return out, err
	}
	return ret, nil
}

// getTargetBranch
// if no parent left, merge it into final branch defined in project config yaml
func (i *Routine) getTargetBranch() string {
	parentBranchName := i.projectCfg.FinalBranch
	parentExists := i.parent != nil && i.parent.Id != 0
	if parentExists {
		parentBranchName = i.workbench.GetIssueBranchName(*i.parent)
	}
	return parentBranchName
}

func (i *Routine) commentCurrentAboutMerge() error {
	parentBranchName := i.getTargetBranch()
	currentBranchName := i.workbench.GetIssueBranchName(i.issue)

	commentText := fmt.Sprintf("Merged #%d branch %q into parent %q", i.issue.Id, currentBranchName, parentBranchName)
	err := i.AddComment(commentText)
	if err != nil {
		return fmt.Errorf("failed to add merge comment to issue: %v", err)
	}
	return nil
}

func (i *Routine) commentParentAboutMerge() error {
	if i.parent == nil || i.parent.Id == 0 {
		return nil
	}

	parentBranchName := i.getTargetBranch()
	currentBranchName := i.workbench.GetIssueBranchName(i.issue)

	branchDiffURL := fmt.Sprintf("[branch diff](/projects/%s/repository/%s/diff?rev=%s&rev_to=%s", i.project.Identifier, i.project.Identifier, parentBranchName, i.projectCfg.FinalBranch)
	commentText := fmt.Sprintf("Merged #%d branch %q into %q. %s)", i.issue.Id, currentBranchName, parentBranchName, branchDiffURL)
	err := i.AddCommentToParent(commentText)
	if err != nil {
		return fmt.Errorf("failed to add merge comment to parent: %v", err)
	}
	return nil
}
