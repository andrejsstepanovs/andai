package employee

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/andrejsstepanovs/andai/internal/exec"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
)

// CommitCommentFormat is the format for the commit comment
const CommitCommentFormat = "%d. Commit [%s](/projects/%s/repository/%s/revisions/%s/diff) - %s"

func (i *Routine) commentCommitsSince(currentCommitSku, commitMessage string) (int, error) {
	newCommits, getShaErr := i.workbench.GetCommitsSinceInReverseOrder(currentCommitSku)

	//commits, getShaErr := i.workbench.GetBranchCommits(10)
	if getShaErr != nil {
		log.Printf("Failed to get last commit sha: %v", getShaErr)
		return 0, fmt.Errorf("failed to get last commit sha: %v", getShaErr)
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
		return 0, nil
	}

	branchName := i.workbench.GetIssueBranchName(i.issue)
	format := "### Branch [%s](/projects/%s/repository/%s?rev=%s)"
	txt := make([]string, 0)
	txt = append(txt, fmt.Sprintf(format, branchName, i.project.Identifier, i.project.Identifier, branchName))
	for n, sha := range newCommits {
		txt = append(txt, fmt.Sprintf(CommitCommentFormat, n+1, sha, i.project.Identifier, i.project.Identifier, sha, commitMessage))
	}

	err := i.AddComment(strings.Join(txt, "\n"))
	if err != nil {
		return len(newCommits), err
	}

	return len(newCommits), nil
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
	_, err = i.commentCommitsSince(lastCommit, commitMessage)
	if err != nil {
		return out, err
	}
	return ret, nil
}

func (i *Routine) parentExists() bool {
	return i.parent != nil && i.parent.Id != 0
}

// getTargetBranch
// if no parent left, merge it into final branch defined in project config yaml
// returning slice in order. Example:
// 1. main
// 2. issue parents parent branch
// 3. issue parent branch
func (i *Routine) getTargetBranch() []string {
	orderedList := make([]string, 0)
	seen := make(map[string]bool)

	if i.parentExists() {
		for _, p := range i.parents {
			if p.Id != 0 {
				branchName := i.workbench.GetIssueBranchName(p)
				if branchName != "" && !seen[branchName] {
					orderedList = append(orderedList, branchName)
					seen[branchName] = true
				}
			}
		}
	}

	// Only add final branch if it's not already present
	if !seen[i.projectCfg.FinalBranch] {
		orderedList = append(orderedList, i.projectCfg.FinalBranch)
	}

	slices.Reverse(orderedList)
	return orderedList
}

func (i *Routine) commentCurrentAboutMerge() error {
	parentBranches := i.getTargetBranch()
	parentBranchName := parentBranches[len(parentBranches)-1]
	currentBranchName := i.workbench.GetIssueBranchName(i.issue)

	commentText := fmt.Sprintf("Merged #%d - Branch %q -> %q", i.issue.Id, currentBranchName, parentBranchName)
	if i.parentExists() {
		commentText = fmt.Sprintf("Merged #%d -> #%d - Branch %q -> %q", i.issue.Id, i.issue.Parent.Id, currentBranchName, parentBranchName)
	}

	err := i.AddComment(commentText)
	if err != nil {
		return fmt.Errorf("failed to add merge comment to issue: %v", err)
	}
	return nil
}

func (i *Routine) commentParentBranchDiff() error {
	if !i.parentExists() {
		return nil
	}

	parentSha := ""
	lastSha := ""
	for _, field := range i.issue.CustomFields {
		if field.Name == model.CustomFieldParentSha {
			parentSha = field.Value.(string)
		}
		if field.Name == model.CustomFieldLastSha {
			lastSha = field.Value.(string)
		}
	}

	if parentSha == "" || lastSha == "" || parentSha == lastSha {
		return nil
	}

	parentBranches := i.getTargetBranch()
	parentBranchName := parentBranches[len(parentBranches)-1]
	currentBranchName := i.workbench.GetIssueBranchName(i.issue)

	branchDiffURL := fmt.Sprintf("[Merge Diff: %s <- %s](/projects/%s/repository/%s/diff?rev=%s&rev_to=%s)", parentBranchName, currentBranchName, i.project.Identifier, i.project.Identifier, lastSha, parentSha)

	existingComments, err := i.getParentComments()
	if err != nil {
		return fmt.Errorf("failed to get parent comments: %v", err)
	}
	for _, comment := range existingComments {
		if comment.Text == branchDiffURL {
			return nil
		}
	}

	err = i.AddCommentToParent(branchDiffURL)
	if err != nil {
		return fmt.Errorf("failed to add merge comment to parent: %v", err)
	}

	{ // mention all in-between commits. will be useful for "context-commits" command later on.
		allLastCommits, err := i.workbench.GetLastCommits(10)
		if err != nil {
			return fmt.Errorf("failed to get last commits: %v", err)
		}

		txt := make([]string, 0)
		start := false
		n := 0
		for _, commit := range allLastCommits {
			if commit == parentSha {
				break
			}
			if commit == lastSha {
				start = true
			}
			if start {
				txt = append(txt, fmt.Sprintf(CommitCommentFormat, n+1, commit, i.project.Identifier, i.project.Identifier, commit, "code changes"))
			}
		}

		if len(txt) == 0 {
			log.Println("No commits to mention in parent branch diff comment")
			return nil
		}
		commentText := fmt.Sprintf("### Commits in branch %q:\n%s\n\n### %s", currentBranchName, strings.Join(txt, "\n"), branchDiffURL)
		err = i.AddComment(commentText)
		if err != nil {
			return fmt.Errorf("failed to add comment about commits in parent branch diff: %v", err)
		}
	}

	return nil
}

func (i *Routine) commentParentAboutMerge() error {
	if !i.parentExists() {
		return nil
	}

	parentBranches := i.getTargetBranch()
	parentBranchName := parentBranches[len(parentBranches)-1]
	currentBranchName := i.workbench.GetIssueBranchName(i.issue)

	commentText := fmt.Sprintf("Merged #%d branch %q into %q", i.issue.Id, currentBranchName, parentBranchName)
	err := i.AddCommentToParent(commentText)
	if err != nil {
		return fmt.Errorf("failed to add merge comment to parent: %v", err)
	}
	return nil
}
