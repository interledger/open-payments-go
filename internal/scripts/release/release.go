package main

import (
	"bytes"
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

	program := "go run internal/scripts/release/release.go"

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
			err = updateGoModAndImports(newVersion.Major)
			if err != nil {
				fmt.Printf("Error: Failed to update 'go.mod': %v\n", err)
				os.Exit(1)
			}
		}
	}

	if needsGoModUpdate {
		if !*dr {
			err = commitAndPush(newVersion.String())
			if err != nil {
				fmt.Printf("Error: Failed to push commit or push changes: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("DRY RUN MODE - Would commit and push changes for %s\n", newVersion)
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

func updateGoModAndImports(newMajor int) error {
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

	files, err := findFilesUsingModule(currentModule)
	if err != nil {
		return fmt.Errorf("failed to find files using module %s: %v", currentModule, err)
	}

	err = updateImportsInFiles(files, currentModule, newModule)
	if err != nil {
		return fmt.Errorf("failed to update imports in files: %v", err)
	}

	return nil
}

func commitAndPush(version string) error {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %v", err)
	}

	if len(output) == 0 {
		fmt.Println("No changes to commit")
		return nil
	}

	fmt.Printf("Detected changes:\n%s\n", output)

	cmd = exec.Command("git", "add", "-u") // We use `-u` to only commit modified files
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git add modified files: %v", err)
	}

	// TODO: Double check to make sure there are not new files and exit with an error code?

	commitMsg := fmt.Sprintf("chore: update module path and related files for %s", version)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	fmt.Printf("Committed changes for %s\n", version)

	cmd = exec.Command("git", "push", "origin", "HEAD")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %v", err)
	}

	fmt.Printf("Pushed changes to remote\n")
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

func findFilesUsingModule(oldModule string) ([]string, error) {
	cmd := exec.Command("grep", "-rl", oldModule, ".")
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("grep error: %s", ee.Stderr)
		}
		return nil, fmt.Errorf("grep failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	fmt.Printf("Found %d files to update.\n", len(lines))
	return lines, nil
}

func updateImportsInFiles(files []string, oldModule, newModule string) error {
	for _, path := range files {
		input, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if !bytes.Contains(input, []byte(oldModule)) {
			continue // sanity check
		}

		output := bytes.ReplaceAll(input, []byte(`"`+oldModule), []byte(`"`+newModule))
		if err := os.WriteFile(path, output, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		fmt.Printf("Updated imports in %s\n", path)
	}
	return nil
}
