package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var cmdRoot = &cobra.Command{
	Use:   "groove",
	Short: "Groove is a HTTP/Cron project scaffold.",
}

var cmdCreate = &cobra.Command{
	Use:     "create PROJECT_NAME",
	Short:   "Create a new Groove project.",
	Example: "groove create my-app",
	Run:     runCmdCreate,
}

func init() {
	cmdRoot.AddCommand(cmdCreate)
}

func main() {
	cmdRoot.Execute()
}

func runCmdCreate(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		log.Print("project name is not given")
		return
	}

	projectName := args[0]
	projectPath := fmt.Sprintf("./%s", projectName)

	_, err := git.PlainCloneContext(context.TODO(), projectPath, false, &git.CloneOptions{
		URL:      "https://github.com/xvrzhao/groove-scaffold.git",
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		log.Printf("failed to pull groove scaffold: %v", err)
		return
	}

	if err := os.RemoveAll(projectPath + "/.git"); err != nil {
		log.Printf("failed to remove .git dir")
		return
	}

	if err := replaceStrRecursive(projectPath, "github.com/xvrzhao/groove-scaffold", projectName); err != nil {
		log.Printf("failed to replace module name: %v", err)
		return
	}

	if _, err := git.PlainInit(projectPath, false); err != nil {
		log.Printf("failed to init git: %v", err)
		return
	}
}

func replaceStrRecursive(root, oldString, newString string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		if err := replaceStrInFile(path, oldString, newString); err != nil {
			return fmt.Errorf("failed to process file %s: %w", path, err)
		}
		return nil
	})
}

func replaceStrInFile(filePath, oldString, newString string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	changed := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, oldString) {
			line = strings.ReplaceAll(line, oldString, newString)
			changed = true
		}
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !changed {
		return nil
	}

	return os.WriteFile(filePath, buffer.Bytes(), 0644)
}
