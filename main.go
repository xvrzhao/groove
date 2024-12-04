package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

const (
	scaffoldRepo    = "https://github.com/xvrzhao/groove-scaffold.git"
	scaffoldGoMod   = "github.com/xvrzhao/groove-scaffold"
	scaffoldVersion = "v1.1.0"
)

var cmdRoot = &cobra.Command{
	Use:   "groove",
	Short: "Groove is a HTTP/Cron project scaffold for singleton.",
}

var cmdCreate = &cobra.Command{
	Use:     "create PROJECT_NAME",
	Short:   "Create a new Groove project",
	Example: "groove create my-app",
	Run:     runCmdCreate,
}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(scaffoldVersion)
	},
}

func init() {
	cmdCreate.Flags().String("d", ".", "directory to create")
	cmdRoot.AddCommand(cmdCreate)
	cmdRoot.AddCommand(cmdVersion)
}

func main() {
	cmdRoot.Execute()
}

func runCmdCreate(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		fmt.Println("project name is not given")
		return
	}

	projectName := args[0]
	dir := cmd.Flag("d").Value.String()
	projectPath, err := filepath.Abs(filepath.Join(dir, projectName))
	if err != nil {
		fmt.Printf("failed to get abs path: %v\n", err)
		return
	}

	if _, err := git.PlainClone(projectPath, false, &git.CloneOptions{
		URL:           scaffoldRepo,
		ReferenceName: scaffoldVersion,
		Depth:         1,
		Progress:      os.Stdout,
	}); err != nil {
		fmt.Printf("failed to pull groove-scaffold code: %v\n", err)
		return
	}

	if err := os.RemoveAll(filepath.Join(projectPath, ".git")); err != nil {
		fmt.Printf("failed to remove git dir: %v\n", err)
		return
	}

	if err := replaceFileStrRecursively(projectPath, scaffoldGoMod, projectName); err != nil {
		fmt.Printf("failed to replace module name: %v\n", err)
		return
	}

	if _, err := git.PlainInit(projectPath, false); err != nil {
		fmt.Printf("failed to init git: %v\n", err)
		return
	}

	fmt.Printf("Complete! Enjoy it: %s\n", projectPath)
}

func replaceFileStrRecursively(dir, oldStr, newStr string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		if err := replaceFileStr(path, oldStr, newStr); err != nil {
			return fmt.Errorf("failed to process file %s: %w", path, err)
		}
		return nil
	})
}

func replaceFileStr(filePath, oldStr, newStr string) error {
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
		if err := scanner.Err(); err != nil {
			return err
		}

		if strings.Contains(line, oldStr) {
			line = strings.ReplaceAll(line, oldStr, newStr)
			changed = true
		}
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	if !changed {
		return nil
	}

	return os.WriteFile(filePath, buffer.Bytes(), 0644)
}
