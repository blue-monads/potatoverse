package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/term"
)

type TemplateInfo struct {
	Name    string
	RepoUrl string
	Slug    string
}

var TEMPLATES = []TemplateInfo{
	{
		Name:    "React Lua",
		RepoUrl: "https://github.com/blue-monads/potato-app-react-lua-template",
		Slug:    "potato-app-react-lua-template",
	},
	{
		Name:    "Vanilla Lua",
		RepoUrl: "https://github.com/blue-monads/potato-app-vannilla-lua-template",
		Slug:    "potato-app-vannilla-lua-template",
	},
}

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

type PackageInitCmd struct {
	Template  string `name:"template" help:"Template slug or name to use. If omitted, choose interactively."`
	Slug      string `name:"slug" help:"Project slug (letters, digits and hyphens only). If omitted, prompt interactively."`
	Directory string `name:"directory" short:"d" help:"Destination directory for new project." type:"path" default:"."`
}

func (c *PackageInitCmd) Run(_ *kong.Context) error {
	templateInfo, err := c.chooseTemplate()
	if err != nil {
		return err
	}

	slug, err := c.chooseSlug()
	if err != nil {
		return err
	}

	destination := filepath.Join(c.Directory, slug)
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	}

	fmt.Printf("\nCreating project %s from template %s...\n", slug, templateInfo.Name)

	if err := cloneTemplateRepo(templateInfo.RepoUrl, destination); err != nil {
		return err
	}

	if err := replaceTemplateSlug(destination, templateInfo.Slug, slug); err != nil {
		return err
	}

	if err := resetGitState(destination); err != nil {
		return err
	}

	fmt.Printf("\nProject created successfully at %s\n", destination)
	fmt.Printf("Next steps:\n  cd %s\n", filepath.Clean(destination))

	return nil
}

func (c *PackageInitCmd) chooseTemplate() (TemplateInfo, error) {
	if c.Template != "" {
		for _, t := range TEMPLATES {
			if strings.EqualFold(t.Slug, c.Template) || strings.EqualFold(t.Name, c.Template) {
				return t, nil
			}
		}
		return TemplateInfo{}, fmt.Errorf("template %q not found", c.Template)
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return TemplateInfo{}, errors.New("interactive template selection requires a TTY or use --template")
	}

	return interactiveTemplateSelect()
}

func (c *PackageInitCmd) chooseSlug() (string, error) {
	if c.Slug != "" {
		if !slugPattern.MatchString(c.Slug) {
			return "", fmt.Errorf("invalid slug %q: only letters, digits, and hyphens are allowed", c.Slug)
		}
		return c.Slug, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("interactive project slug prompt requires a TTY or use --slug")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Project slug: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		slug := strings.TrimSpace(input)
		if slug == "" {
			fmt.Println("Slug is required. Use letters, digits and hyphens only.")
			continue
		}
		if !slugPattern.MatchString(slug) {
			fmt.Println("Invalid slug. Only letters, digits and hyphens are allowed.")
			continue
		}
		return slug, nil
	}
}

func interactiveTemplateSelect() (TemplateInfo, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return TemplateInfo{}, err
	}
	defer term.Restore(fd, oldState)

	reader := bufio.NewReader(os.Stdin)
	selection := 0
	renderTemplateMenu(selection)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return TemplateInfo{}, err
		}

		if b == '\r' || b == '\n' {
			fmt.Print("\n")
			return TEMPLATES[selection], nil
		}

		if b == '\x03' {
			return TemplateInfo{}, errors.New("aborted")
		}

		if b != '\x1b' {
			continue
		}

		second, err := reader.ReadByte()
		if err != nil {
			return TemplateInfo{}, err
		}
		if second != '[' {
			continue
		}

		third, err := reader.ReadByte()
		if err != nil {
			return TemplateInfo{}, err
		}

		if third == 'A' {
			selection--
			if selection < 0 {
				selection = len(TEMPLATES) - 1
			}
			renderTemplateMenu(selection)
			continue
		}
		if third == 'B' {
			selection++
			if selection >= len(TEMPLATES) {
				selection = 0
			}
			renderTemplateMenu(selection)
			continue
		}
	}
}

func renderTemplateMenu(selection int) {
	fmt.Print("\x1b[2J\x1b[H") // clear screen and home cursor
	fmt.Println("Select a template with the arrow keys and press Enter:")
	for idx, template := range TEMPLATES {
		if idx == selection {
			fmt.Printf("\x1b[7m> %s\x1b[0m\n", template.Name)
		} else {
			fmt.Printf("  %s\n", template.Name)
		}
	}
}

func cloneTemplateRepo(repoUrl, destination string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", repoUrl, destination)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func replaceTemplateSlug(root, oldSlug, newSlug string) error {
	oldBytes := []byte(oldSlug)
	newBytes := []byte(newSlug)
	pathsToRename := make([]string, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}

		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Contains(data, oldBytes) {
				info, err := os.Stat(path)
				if err != nil {
					return err
				}
				newData := bytes.ReplaceAll(data, oldBytes, newBytes)
				if err := os.WriteFile(path, newData, info.Mode()); err != nil {
					return err
				}
			}
		}

		if strings.Contains(d.Name(), oldSlug) {
			pathsToRename = append(pathsToRename, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(pathsToRename, func(i, j int) bool {
		return len(pathsToRename[i]) > len(pathsToRename[j])
	})

	for _, oldPath := range pathsToRename {
		base := filepath.Base(oldPath)
		newBase := strings.ReplaceAll(base, oldSlug, newSlug)
		newPath := filepath.Join(filepath.Dir(oldPath), newBase)
		if oldPath == newPath {
			continue
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return nil
}

func resetGitState(root string) error {
	gitDir := filepath.Join(root, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
