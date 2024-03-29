package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/mod/modfile"
)

type Upgrader struct {
	repo                      *git.Repository
	rootDir                   string
	currentModuleVersions     map[string]string
	currentModuleDependencies map[string]map[string]string
}

func (u *Upgrader) loadModuleDependencies() error {
	for moduleName, currentVersion := range u.currentModuleVersions {
		moduleLocalPath := strings.TrimPrefix(moduleName, "github.com/go-pnp/go-pnp")
		if strings.HasPrefix(moduleLocalPath, "/") {
			moduleLocalPath = moduleLocalPath[1:]
		}
		tagName := moduleLocalPath + "/v" + currentVersion
		if moduleLocalPath == "" {
			tagName = "v" + currentVersion
		}
		tag, err := u.repo.Tag(moduleLocalPath + "/v" + currentVersion)
		if err != nil {
			return fmt.Errorf("tag '%s': %w", tagName, err)
		}

		var tree *object.Tree
		tagObject, err := u.repo.TagObject(tag.Hash())
		if err != nil {
			commitObject, err := u.repo.CommitObject(tag.Hash())
			if err != nil {
				return fmt.Errorf("commit object '%s': %w", tagName, err)
			}
			tree, err = commitObject.Tree()
			if err != nil {
				return fmt.Errorf("object type: %w", err)
			}
		} else {
			tree, err = tagObject.Tree()
			if err != nil {
				return fmt.Errorf("object type: %w", err)
			}
		}

		modFile := filepath.Join(moduleLocalPath, "go.mod")
		modTreeFile, err := tree.File(modFile)
		if err != nil {
			return fmt.Errorf("file %s: %w", modFile, err)
		}

		modBlob, err := modTreeFile.Contents()
		if err != nil {
			return fmt.Errorf("contents: %w", err)
		}
		modfile, err := modfile.Parse("go.mod", []byte(modBlob), nil)
		if err != nil {
			return fmt.Errorf("parse go.mod: %w", err)
		}

		for _, require := range modfile.Require {
			if !strings.HasPrefix(require.Mod.Path, "github.com/go-pnp/go-pnp") {
				continue
			}

			if _, ok := u.currentModuleDependencies[moduleName]; !ok {
				u.currentModuleDependencies[moduleName] = make(map[string]string)
			}
			u.currentModuleDependencies[moduleName][require.Mod.Path] = strings.TrimPrefix(require.Mod.Version, "v")
		}
	}

	return nil
}

func (u *Upgrader) loadModuleVersions() error {
	allModuleVersions := make(map[string][]string)
	tagsIter, err := u.repo.Tags()
	if err != nil {
		return fmt.Errorf("tags: %w", err)
	}

	err = tagsIter.ForEach(func(ref *plumbing.Reference) error {
		tagName := strings.TrimPrefix(ref.Name().String(), "refs/tags/")

		tagParts := strings.Split(tagName, "/")
		if len(tagParts) < 2 {
			if strings.HasPrefix(tagName, "v") {
				moduleName := ""
				version := strings.TrimPrefix(tagName, "v")
				if _, ok := allModuleVersions[moduleName]; !ok {
					allModuleVersions[moduleName] = make([]string, 0, 1)
				}

				allModuleVersions[moduleName] = append(allModuleVersions[moduleName], version)
			}
			return nil
		}

		moduleName := strings.Join(tagParts[0:len(tagParts)-1], "/")
		version := tagParts[len(tagParts)-1]
		if !strings.HasPrefix(version, "v") {
			return nil
		}

		if _, ok := allModuleVersions[moduleName]; !ok {
			allModuleVersions[moduleName] = make([]string, 0, 1)
		}

		allModuleVersions[moduleName] = append(allModuleVersions[moduleName], strings.TrimPrefix(version, "v"))
		return nil
	})

	u.currentModuleVersions = make(map[string]string)
	for moduleName, versions := range allModuleVersions {
		sort.Sort(sort.StringSlice(versions))
		if moduleName == "" {
			u.currentModuleVersions["github.com/go-pnp/go-pnp"] = versions[len(versions)-1]
		} else {
			u.currentModuleVersions["github.com/go-pnp/go-pnp/"+moduleName] = versions[len(versions)-1]
		}
	}

	return nil
}

func (u *Upgrader) Release() {
	updatedPaths := make(map[string]struct{})
	for moduleName, dependencies := range u.currentModuleDependencies {
		for dependency, version := range dependencies {
			if version != u.currentModuleVersions[dependency] {
				fmt.Printf("module %s depends on %s@%s, but the latest version is %s\n", moduleName, dependency, version, u.currentModuleVersions[dependency])
				moduleLocalPath := strings.TrimPrefix(moduleName, "github.com/go-pnp/go-pnp")
				if strings.HasPrefix(moduleLocalPath, "/") {
					moduleLocalPath = moduleLocalPath[1:]
				}
				// executing go get
				cmd := exec.Command("go", "get", "-u", dependency+"@v"+u.currentModuleVersions[dependency])
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Dir = filepath.Join(u.rootDir, moduleLocalPath)
				if err := cmd.Run(); err != nil {
					fmt.Printf("can't upgrade module %s: %v\n", moduleLocalPath, err)
				}

				// executing go mod tidy
				cmd = exec.Command("go", "mod", "tidy")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Dir = filepath.Join(u.rootDir, moduleLocalPath)
				if err := cmd.Run(); err != nil {
					fmt.Printf("can't tidy module %s: %v\n", moduleLocalPath, err)
				}
				updatedPaths[moduleLocalPath] = struct{}{}
			}
		}
	}

	w, err := u.repo.Worktree()
	if err != nil {
		fmt.Printf("can't get worktree: %v\n", err)
		return
	}

	for path := range updatedPaths {
		_, err = w.Add(filepath.Join(path, "go.mod"))
		if err != nil {
			fmt.Printf("can't add %s: %v\n", path, err)
		}
		_, err = w.Add(filepath.Join(path, "go.sum"))
		if err != nil {
			fmt.Printf("can't add %s: %v\n", path, err)
		}
	}
	status, err := w.Status()
	if err != nil {
		fmt.Printf("can't get status: %v\n", err)
		return
	}

	var hasChanges bool
	for path := range updatedPaths {
		if status[filepath.Join(path, "go.mod")].Staging != git.Unmodified || status[filepath.Join(path, "go.sum")].Staging != git.Unmodified {
			hasChanges = true
			break
		}
	}
	if !hasChanges {
		fmt.Println("no changes to commit")
		return
	}

	commit, err := w.Commit("upgrade modules", &git.CommitOptions{})
	if err != nil {
		fmt.Printf("can't commit: %v\n", err)
		return
	}
	//
	//fmt.Printf("commit %s\n", commit.String())
	fmt.Println("comitting", updatedPaths)
	// tag for modules
	for updatedModule := range updatedPaths {
		tag := u.currentModuleVersions["github.com/go-pnp/go-pnp/"+updatedModule]
		// increase minor version
		tagParts := strings.Split(tag, ".")
		minorVersion, err := strconv.ParseInt(tagParts[2], 10, 64)
		if err != nil {
			fmt.Printf("can't parse minor version: %v\n", err)
			return
		}
		tag = fmt.Sprintf("%s/v%s.%s.%d", updatedModule, tagParts[0], tagParts[1], minorVersion+1)
		fmt.Println("tagging with ", tag)
		_, err = u.repo.CreateTag(tag, commit, &git.CreateTagOptions{
			Message: "sync module dependencies",
		})
		if err != nil {
			fmt.Printf("can't create tag %s: %v\n", tag, err)
			return
		}
		fmt.Printf("tag %s created\n", tag)
	}
}

func newUpgrader(rootDir string) (*Upgrader, error) {
	repo, err := git.PlainOpen(rootDir)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	result := &Upgrader{
		repo:                      repo,
		rootDir:                   rootDir,
		currentModuleVersions:     make(map[string]string),
		currentModuleDependencies: make(map[string]map[string]string),
	}

	if err := result.loadModuleVersions(); err != nil {
		return nil, fmt.Errorf("load module versions: %w", err)
	}

	if err := result.loadModuleDependencies(); err != nil {
		return nil, fmt.Errorf("load module dependencies: %w", err)
	}

	return result, nil
}

func main() {
	upgrader, err := newUpgrader("..")
	if err != nil {
		fmt.Println("can't create upgrader:", err)
		os.Exit(1)
	}
	upgrader.Release()
}
