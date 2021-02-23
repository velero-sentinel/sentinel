// +build mage

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

type Docker mg.Namespace

const repoUrl = "https://github.com/velero-sentinel/sentinel"
const imageName = "velerosentinel/sentinel"

type Sonarcloud mg.Namespace

// Creates the coverage output for SonarCloud
func (Sonarcloud) Coverage() error {
	jsonraw, err := sh.Output("go", "test", "-coverprofile=cover.out", "-json", "./...")

	if err != nil {
		return fmt.Errorf("creating json report: %s", err)
	}
	return ioutil.WriteFile("./testreport.json", []byte(jsonraw), 0644)
}

// Creates the report required for SonarCloud
func (Sonarcloud) Reports() {
	// We should have a look at
	mg.Deps(Sonarcloud.Coverage)
}

func rv() (rev, version string) {
	var err error
	if rev, err = sh.Output("git", "rev-parse", "--short", "HEAD"); err != nil {
		log.Fatalf("obtaining git hash: %s", err)
	}
	var tag string
	if tag, err = sh.Output("git", "rev-list", "--tags", "--max-count=1"); err != nil {
		log.Fatalf("obtaining tag revision: %s", err)
	}

	if version, err = sh.Output("git", "describe", "--tags", tag); err != nil {
		log.Fatalf("obtaining version: %s", err)
	}

	return
}

// Create the docker image
func (Docker) Build(ctx context.Context) error {

	mg.Deps(Test.Test)
	rev, version := rv()

	date := time.Now().Format(time.RFC3339)

	return sh.RunV("docker", "build",
		"-t", fmt.Sprintf("%s:%s", imageName, version),
		"-t", fmt.Sprintf("%s:latest", imageName),
		"--build-arg", fmt.Sprintf("REPO_URL=%s", repoUrl),
		"--build-arg", fmt.Sprintf("GIT_REVISION=%s", rev),
		"--build-arg", fmt.Sprintf("VERSION=%s", version),
		"--build-arg", fmt.Sprintf("BUILD_DATE=%s", date),
		".",
	)
}

type Test mg.Namespace

// Create a code coverage profile
func (Test) Coverage() error {
	return sh.RunV("go", "test", "-v", "-coverpkg=./...", "-coverprofile=cover.out", "./...")
}

// Create the HTML coverage report
func (Test) Html() error {
	mg.Deps(Test.Coverage)
	return sh.Run("go", "tool", "cover", "-html", "cover.out")
}

// Runs the tests
func (Test) Test() error {
	return sh.RunV("go", "test", "-v", "./...")
}

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// Builds the apllication
func Build() error {
	fmt.Println("Building...")
	rev, version := rv()
	cmd := exec.Command("go", "build", "-ldflags",
		fmt.Sprintf("-X main.GitCommit=%s -X main.Version=%s -X main.BuildDate=%s", rev, version, time.Now().Format(time.RFC3339)))
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	for _, f := range []string{"sentinel", "testreport.json", "cover.out"} {
		os.Remove(f)
	}
}
