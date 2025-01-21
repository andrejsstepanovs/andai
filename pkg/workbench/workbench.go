package workbench

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andrejsstepanovs/andai/pkg/worker"
	"github.com/mattn/go-redmine"
)

type Workbench struct {
	Git        *worker.Git
	Issue      redmine.Issue
	WorkingDir string
}

type git interface {
	CheckoutBranch(name string) error
	GetPath() string
}

func (i *Workbench) PrepareWorkplace() error {
	err := i.changeDirectory()
	if err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}

	err = i.checkoutBranch()
	if err != nil {
		log.Printf("Failed to checkout branch: %v", err)
		return err
	}

	return nil
}

func (i *Workbench) changeDirectory() error {
	targetPath := i.Git.GetPath()
	if filepath.Base(targetPath) == ".git" {
		targetPath = filepath.Dir(targetPath)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory err: %v", err)
	}
	if currentDir != targetPath {
		log.Printf("Changing directory from %s to %s\n", currentDir, targetPath)
	}

	err = os.Chdir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to change directory err: %v", err)
	}

	log.Printf("Active in project directory %s\n", targetPath)
	i.WorkingDir = targetPath

	return nil
}

func (i *Workbench) checkoutBranch() error {
	err := i.Git.CheckoutBranch(strconv.Itoa(i.Issue.Id))
	if err != nil {
		return fmt.Errorf("failed to checkout branch err: %v", err)
	}
	return nil
}
