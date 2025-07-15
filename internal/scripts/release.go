package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type version struct {
	Major, Minor, Patch int
}

func (v version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

type BumpType string

const (
	patch BumpType = "patch"
	minor BumpType = "minor"
	major BumpType = "major"
)

func (b BumpType) IsValid() bool {
	return b == patch || b == minor || b == major
}

func main() {
	var (
		bt = flag.String("type", "", "Version bump type: major, minor, or patch")
		dr = flag.Bool("dry-run", false, "Show what would be done without making changes")
	)

	program := "go run internal/scripts/release.go"

	flag.Usage = func() {
		fmt.Printf("Usage: %s -type=<bump_type>\n\n", program)
		fmt.Printf("Options:\n")
		flag.PrintDefaults()
		fmt.Printf("\nExamples:\n")
		fmt.Printf("  %s -type=patch     # Bump patch version (1.0.0 -> 1.0.1)\n", program)
		fmt.Printf("  %s -type=minor     # Bump minor version (1.0.0 -> 1.1.0)\n", program)
		fmt.Printf("  %s -type=major     # Bump major version (1.0.0 -> 2.0.0)\n", program)
		fmt.Printf("  %s -type=patch -dry-run  # Show what would happen\n", program)
	}

	flag.Parse()

	if *bt == "" {
		fmt.Printf("Error: -type flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	bump := BumpType(*bt)
	if !bump.IsValid() {
		fmt.Printf("Error: Invalid bump type '%s'. Must be 'major', 'minor', or 'patch'\n", *bt)
		os.Exit(1)
	}

	if *dr {
		fmt.Println("DRY RUN MODE - No changes will be made")
	}

	currentVersion, err := getCurrentVersion()
	if err != nil {
		fmt.Printf("Error: Could not retrieve current version: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Current version: %s\n", currentVersion)

	newVersion := bumpVersion(currentVersion, bump)
	fmt.Printf("New version: %s\n", newVersion)

	needsGoModUpdate := bump == major && currentVersion.Major >= 0

	if needsGoModUpdate {
		fmt.Printf("Major version bump detected - 'go.mod' needs update\n")
		if !*dr {
			err = updateGoMod(newVersion.Major)
			if err != nil {
				fmt.Printf("Error: Failed to update 'go.mod': %v\n", err)
				os.Exit(1)
			}
		}
	}

	// I am still not sure if this is the correct way of doing this. My though process:
	//
	// 1. If we want to release a new major version, update the go.mod file by
	//    appending/increasing `/v${MAJOR_VERSION}` in the module name.
	// 2. Push the updated 'go.mod' file to GitHub.
	// 3. Tag & push
	if needsGoModUpdate {
		if !*dr {
			err = commitAndPushGoModChanges(newVersion.String())
			if err != nil {
				fmt.Printf("Error: Failed to push commit or push go.mod changes: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("DRY RUN MODE - Would commit and push go.mod changes for %s\n", newVersion)
		}
	}

	if !*dr {
		err = createAndPushTag(newVersion.String())
		if err != nil {
			fmt.Printf("Error: Failed to push tag: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("DRY RUN MODE - Would create and push tag: %s\n", newVersion)
	}

	if *dr {
		fmt.Printf("DRY RUN MODE - Complete! Would release %s\n", newVersion)
	}

	if needsGoModUpdate && !*dr {
		fmt.Printf("Module path updated for major version bump\n")
	}
}

func getCurrentVersion() (version, error) {
	cmd := exec.Command("git", "tag", "-l", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: Could not list already existing tags: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		if version, err := parseVersion(line); err == nil {
			return version, nil
		}
	}

	return version{0, 0, 0}, nil
}

func parseVersion(tag string) (version, error) {
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) != 4 {
		return version{}, fmt.Errorf("invalid version format: %s", tag)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return version{major, minor, patch}, nil
}

func bumpVersion(current version, bumpType BumpType) version {
	switch bumpType {
	case major:
		return version{current.Major + 1, 0, 0}
	case minor:
		return version{current.Major, current.Minor + 1, 0}
	case patch:
		return version{current.Major, current.Minor, current.Patch + 1}
	default:
		return current
	}
}

func updateGoMod(newMajor int) error {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	currentModule := strings.TrimSpace(string(output))

	re := regexp.MustCompile(`/v\d+$`)
	baseModule := re.ReplaceAllString(currentModule, "")

	var newModule string
	if newMajor >= 2 {
		newModule = fmt.Sprintf("%s/v%d", baseModule, newMajor)
	} else {
		newModule = baseModule
	}

	fmt.Printf("Updating module path: %s -> %s\n", currentModule, newModule)

	cmd = exec.Command("go", "mod", "edit", "-module="+newModule)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update go.mod: %v", err)
	}

	cmd = exec.Command("go", "mod", "tidy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %v", err)
	}

	return nil
}

func commitAndPushGoModChanges(version string) error {
	cmd := exec.Command("git", "diff", "--quiet", "go.mod", "go.sum")
	if err := cmd.Run(); err == nil {
		fmt.Println("No go.mod changes to commit")
		return nil
	}

	cmd = exec.Command("git", "add", "go.mod", "go.sum")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add go.mod/go.sum: %v", err)
	}

	commitMsg := fmt.Sprintf("chore: update module path for %s", version)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit go.mod changes: %v", err)
	}

	fmt.Printf("Committed go.mod changes for %s\n", version)

	cmd = exec.Command("git", "push", "origin", "HEAD")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push go.mod changes: %v", err)
	}

	fmt.Printf("Pushed go.mod changes to remote\n")
	return nil
}

func createAndPushTag(version string) error {
	cmd := exec.Command("git", "tag", version)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}

	fmt.Printf("Created tag: %s\n", version)

	cmd = exec.Command("git", "push", "origin", version)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push tag: %v", err)
	}

	fmt.Printf("Pushed tag: %s\n", version)
	return nil
}
