package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/blue-monads/potatoverse/docs"
)

type SkillsCmd struct {
	List   SkillsListCmd   `cmd:"" help:"List available skills."`
	Show   SkillsShowCmd   `cmd:"" help:"Show skill file content."`
	Export SkillsExportCmd `cmd:"" help:"Export all skills to local directory."`
}

type SkillsListCmd struct{}

func (c *SkillsListCmd) Run() error {
	err := fs.WalkDir(docs.Skills, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Printf("dir:  %s\n", path)
		} else {
			fmt.Printf("file: %s\n", path)
		}
		return nil
	})
	return err
}

type SkillsShowCmd struct {
	Path string `arg:"" help:"Path to the file to show (relative to skills root)."`
}

func (c *SkillsShowCmd) Run() error {
	// Ensure the path starts with skills/ if it doesn't
	fullPath := c.Path
	if !strings.HasPrefix(fullPath, "skills/") {
		fullPath = "skills/" + fullPath
	}

	data, err := fs.ReadFile(docs.Skills, fullPath)
	if err != nil {
		return fmt.Errorf("could not read file %s: %v", fullPath, err)
	}

	fmt.Println(string(data))
	return nil
}

type SkillsExportCmd struct {
	Dest string `arg:"" help:"Destination directory." default:"."`
}

func (c *SkillsExportCmd) Run() error {
	fmt.Printf("Exporting skills to %s...\n", c.Dest)

	err := fs.WalkDir(docs.Skills, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Relativize path by removing "skills/" prefix
		relPath, err := filepath.Rel("skills", path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(c.Dest, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read from embedded FS and write to local disk
		srcFile, err := docs.Skills.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})

	if err == nil {
		fmt.Println("Export completed successfully.")
	}
	return err
}
