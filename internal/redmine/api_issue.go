package redmine

import (
	"fmt"
	"log"
	"sort"

	"github.com/andrejsstepanovs/andai/internal/settings"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

const (
	// RelationBlocks is a relation type that blocks the issue
	RelationBlocks = "blocks"
)

// APIGetChildren returns first level children of the issue. Children are not dependencies on each other.
func (c *Model) APIGetChildren(issue redmine.Issue) ([]redmine.Issue, error) {
	projectIssues, err := c.API().IssuesOf(issue.Project.Id)
	if err != nil {
		return nil, fmt.Errorf("error redmine issues of project: %v", err)
	}

	children := make([]redmine.Issue, 0)
	for _, projectIssue := range projectIssues {
		if projectIssue.Id == issue.Id {
			continue
		}
		if projectIssue.Parent == nil {
			continue
		}
		if projectIssue.Parent.Id == issue.Id {
			children = append(children, projectIssue)
		}
	}

	children = sortDescID(children)
	return children, nil
}

func (c *Model) APIGetParent(issue redmine.Issue) (parent *redmine.Issue, err error) {
	if issue.Parent != nil && issue.Parent.Id != 0 {
		parent, err = c.API().Issue(issue.Parent.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get redmine parent issue err: %v", err)
		}
		//log.Printf("Parent Issue %d: %s", parent.Id, parent.Subject)
	}
	return parent, nil
}

func (c *Model) APIGetAllParents(issue redmine.Issue) ([]redmine.Issue, error) {
	var parents []redmine.Issue
	maxDeep := 50
	for {
		maxDeep--
		if maxDeep == 0 {
			return nil, fmt.Errorf("max deep reached")
		}
		parent, err := c.APIGetParent(issue)
		if err != nil {
			return nil, fmt.Errorf("failed to get redmine parent issue err: %v", err)
		}
		if parent == nil {
			break
		}
		parents = append(parents, *parent)
		issue = *parent
	}

	parents = sortDescID(parents)
	return parents, nil
}

func sortDescID(parents []redmine.Issue) []redmine.Issue {
	sort.SliceIsSorted(parents, func(i, j int) bool {
		return parents[i].Id > parents[j].Id
	})
	return parents
}

func (c *Model) APIGetProjectIssues(project redmine.Project) ([]redmine.Issue, error) {
	projectIssues, err := c.API().IssuesOf(project.Id)
	if err != nil {
		return nil, fmt.Errorf("error redmine issues of project: %v", err)
	}
	return projectIssues, nil
}

func (c *Model) GetValidProjects() ([]redmine.Project, error) {
	allProjects, err := c.APIGetProjects()
	if err != nil {
		return nil, fmt.Errorf("error redmine issue status: %v", err)
	}

	projects := make([]redmine.Project, 0)
	for _, project := range allProjects {
		isPublic, err := c.DBProjectPublic(project.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get redmine project public err: %v", err)
		}

		if !isPublic {
			log.Printf("Skipping not public project %q", project.Identifier)
			continue
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (c *Model) APIGetWorkableIssues(workflow settings.Workflow, projects []redmine.Project) ([]redmine.Issue, error) {
	for _, project := range projects {
		//log.Printf("Project %q\n", project.Identifier)
		activeProjectIssues, err := c.APIGetProjectIssues(project)
		if err != nil {
			return nil, fmt.Errorf("error redmine issues of project: %v", err)
		}

		//for _, issue := range activeProjectIssues {
		//	log.Printf("Issue %d: %s\n", issue.Id, issue.Subject)
		//}

		// dependencies can contain closed issue ids
		dependencies, err := c.issueDependencies(activeProjectIssues)
		if err != nil {
			return nil, fmt.Errorf("error redmine issue dependencies: %v", err)
		}

		cleanedDependencies := c.removeClosedDependencies(dependencies, activeProjectIssues)
		//for issueID, blockedByIDs := range cleanedDependencies {
		//	if len(blockedByIDs) == 0 {
		//		log.Printf("Issue (%d) is not blocked at all\n", issueID)
		//		continue
		//	}
		//	blocked := make([]string, 0)
		//	for _, blockedBy := range blockedByIDs {
		//		blocked = append(blocked, strconv.Itoa(blockedBy))
		//	}
		//	log.Printf("Issue (%d) BLOCKED BY: %v\n", issueID, strings.Join(blocked, ", "))
		//}

		//for issueID, blockedByIDs := range cleanedDependencies {
		//	if len(blockedByIDs) == 0 {
		//		log.Printf("Issue (%d) is not blocked at all\n", issueID)
		//		continue
		//	}
		//
		//	blocked := make([]string, 0)
		//	for _, blockedBy := range blockedByIDs {
		//		blocked = append(blocked, strconv.Itoa(blockedBy))
		//	}
		//	log.Printf("Issue (%d) BLOCKED BY: %v\n", issueID, strings.Join(blocked, ", "))
		//}

		unblockedIDs := make([]int, 0)
		for issueID, depIDs := range cleanedDependencies {
			if len(depIDs) == 0 {
				unblockedIDs = append(unblockedIDs, issueID)
			}
		}

		if len(unblockedIDs) == 0 {
			//log.Printf("No workable issues for %q", project.Identifier)
			continue
		}
		//ids := make([]string, 0)
		//for _, id := range unblockedIDs {
		//	ids = append(ids, fmt.Sprintf("%d", id))
		//}
		//fmt.Printf("UNBLOCKED ISSUES (%d) %s\n", len(unblockedIDs), strings.Join(ids, ", "))

		validIssues := make([]redmine.Issue, 0)
		for _, okIssueID := range unblockedIDs {
			for _, issue := range activeProjectIssues {
				if issue.Id == okIssueID {
					validIssues = append(validIssues, issue)
				}
			}
		}

		//fmt.Printf("VALID ISSUES TO CHECK PRIORITIES AGAINST (%d)\n", len(validIssues))
		//for _, issue := range validIssues {
		//	fmt.Println("ISSUE:", issue.Tracker.Name, issue.Id)
		//}

		issues := c.getCorrectIssue(validIssues, workflow.Priorities)
		if len(issues) > 0 {
			//log.Println("Found workable issues:")
			//for _, issue := range issues {
			//	log.Println(" > ", issue.Id, issue.Subject)
			//}
			return issues, nil
		}
	}

	return nil, nil
}

func (c *Model) removeClosedDependencies(dependencies map[int][]int, issues []redmine.Issue) map[int][]int {
	cleaned := make(map[int][]int)
	for issueID, blockedByIDs := range dependencies {
		if len(blockedByIDs) == 0 {
			cleaned[issueID] = make([]int, 0)
			continue
		}
		cleaned[issueID] = make([]int, 0)

		for _, blockedBy := range blockedByIDs {
			// if dont exist then its closed
			isClosed := true
			for _, issue := range issues {
				if issue.Id == blockedBy {
					isClosed = false
					break
				}
			}

			if !isClosed {
				cleaned[issueID] = append(cleaned[issueID], blockedBy)
			}
		}
	}

	return cleaned
}

func (c *Model) SetBlocksDependency(issueID int, dependencyID int) error {
	issueRelation := redmine.IssueRelation{
		IssueId:      issueID,
		IssueToId:    dependencyID,
		RelationType: RelationBlocks,
	}
	_, err := c.API().CreateIssueRelation(issueRelation)
	if err != nil {
		return fmt.Errorf("error redmine issue relation: %v", err)
	}
	return nil
}

func (c *Model) issueDependencies(projectIssues []redmine.Issue) (map[int][]int, error) {
	dependencies := make(map[int][]int)
	for _, issue := range projectIssues {
		dependencies[issue.Id] = make([]int, 0)
		relations, err := c.API().IssueRelations(issue.Id)
		if err != nil && err.Error() != "Not Found" {
			return dependencies, err
		}
		//log.Printf("Issue (%d) Relations: %d\n", issue.Id, len(relations))
		for _, relation := range relations {
			if relation.RelationType != RelationBlocks {
				continue
			}
			//log.Printf("Issue (%d) - %d is blocked by %d <- needs to be done first\n", issue.Id, relation.IssueToId, relation.IssueId)
			dependencies[relation.IssueToId] = append(dependencies[relation.IssueId], relation.IssueId)
		}
	}

	return dependencies, nil
}

func (c *Model) getCorrectIssue(issues []redmine.Issue, priorities settings.Priorities) []redmine.Issue {
	valid := make([]redmine.Issue, 0)
	for _, priority := range priorities {
		//log.Printf("%d PRIORITY: %q @ %q\n", i, priority.Type, priority.State)
		//fmt.Printf("PRIORITY: %q @ %q\n", priority.Type, priority.State)

		for _, issue := range issues {
			//fmt.Printf("ISSUE: %q (%d) - %q\n", issue.Tracker.Name, issue.Id, issue.Status.Name)
			if issue.Tracker.Name != string(priority.Type) {
				//fmt.Printf("SKIP %q (%d) - not %q\n", issue.Tracker.Name, issue.Id, priority.Type)
				continue
			}
			if issue.Status.Name != string(priority.State) {
				//fmt.Printf("SKIP %q (%d) - %q != %q\n", issue.Tracker.Name, issue.Id, issue.Status.Name, priority.State)
				continue
			}
			//fmt.Println("ISSUE:", issue.Tracker.Name, issue.Id)
			valid = append(valid, issue)
		}
	}

	return valid
}

func (c *Model) Comment(issue redmine.Issue, text string) error {
	issue.Notes = text
	err := c.API().UpdateIssue(issue)
	if err != nil {
		return fmt.Errorf("error redmine issue comment: %v", err)
	}
	return nil
}

func (c *Model) Transition(issue redmine.Issue, nextStatus redmine.IssueStatus) error {
	issue.StatusId = nextStatus.Id

	err := c.API().UpdateIssue(issue)
	if err != nil {
		return fmt.Errorf("error redmine issue comment: %v", err)
	}
	return nil
}

func (c *Model) CreateIssue(issue redmine.Issue) (redmine.Issue, error) {
	created, err := c.API().CreateIssue(issue)
	if err != nil {
		return redmine.Issue{}, fmt.Errorf("error redmine issue comment: %v", err)
	}
	return *created, nil
}

func (c *Model) CreateChildIssuesWithDependencies(trackerID int, parent redmine.Issue, issues map[int]redmine.Issue, deps map[int][]int) error {
	createdIDs := make(map[int]int)
	for k, issue := range issues {
		projectID := parent.Project.Id
		issue.ProjectId = projectID
		issue.Project = &redmine.IdName{Id: projectID}
		issue.ParentId = parent.Id
		issue.Parent = &redmine.Id{Id: parent.Id}
		issue.TrackerId = trackerID

		created, err := c.CreateIssue(issue)
		if err != nil {
			log.Printf("Failed to create issue: %v", err)
			return err
		}
		log.Printf("Created issue: %d\n", created.Id)
		createdIDs[k] = created.Id
	}

	for k, issueID := range createdIDs {
		for _, depK := range deps[k] {
			if createdIDs[depK] == issueID {
				continue
			}
			err := c.DBCreateBlockedByIssueRelation(issueID, createdIDs[depK])
			if err != nil {
				log.Printf("Failed to set blocks dependency: %v", err)
				return err
			}
		}
	}

	return nil
}
