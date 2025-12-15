package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	imageExtensions = []string{"*.jpg", "*.JPG", "*.jpeg", "*.JPEG"}
	videoExtensions = []string{"*.mov", "*.MOV"}
)

// FileOrganiser defines the interface for organising files
type FileOrganiser interface {
	// OrganiseByDate moves files to date-based directories
	OrganiseByDate(sourceDir, targetDir string) error
	// OrganiseVideosAndRenameJPGs organises videos into subdirectories and renames JPGs sequentially
	OrganiseVideosAndRenameJPGs(targetDir string) error
	// RenameDirectory renames a date-based directory and all images inside it
	RenameDirectory(directory, newName string) error
}

// fileOrganiser implements the FileOrganiser interface
type fileOrganiser struct{}

// NewFileOrganiser creates a new FileOrganiser instance
func NewFileOrganiser() FileOrganiser {
	return &fileOrganiser{}
}

// OrganiseByDate moves files to date-based directories
func (o *fileOrganiser) OrganiseByDate(sourceDir, targetDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(sourceDir, entry.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		dirName := info.ModTime().Format("2006 01 January 02")
		destDir := filepath.Join(targetDir, dirName)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}
		if err := os.Rename(filePath, filepath.Join(destDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// OrganiseVideosAndRenameJPGs organises videos into subdirectories and renames JPGs sequentially
func (o *fileOrganiser) OrganiseVideosAndRenameJPGs(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(targetDir, entry.Name())
		if err := o.organiseVideos(dirPath, entry.Name()); err != nil {
			return err
		}
		if err := o.renameJPGs(dirPath, entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

// organiseVideos moves MOV files to a videos subdirectory and renames them sequentially
func (o *fileOrganiser) organiseVideos(dir string, dirName string) error {
	videoFiles := []string{}
	for _, pattern := range videoExtensions {
		files, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return err
		}
		videoFiles = append(videoFiles, files...)
	}
	if len(videoFiles) == 0 {
		return nil
	}
	videosDir := filepath.Join(dir, "videos")
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		return err
	}
	sort.Strings(videoFiles)
	parts := strings.Fields(dirName)
	if len(parts) != 4 {
		return fmt.Errorf("unexpected directory name format: %s", dirName)
	}
	videosName := strings.Join(parts, "_")
	for i, file := range videoFiles {
		ext := filepath.Ext(file)
		newPath := filepath.Join(videosDir, fmt.Sprintf("%s_%05d%s", videosName, i+1, ext))
		if err := os.Rename(file, newPath); err != nil {
			return err
		}
	}
	return nil
}

// renameJPGs renames JPG files with a sequential pattern
func (o *fileOrganiser) renameJPGs(dir, dirName string) error {
	imageFiles := []string{}
	for _, pattern := range imageExtensions {
		files, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return err
		}
		imageFiles = append(imageFiles, files...)
	}
	if len(imageFiles) == 0 {
		return nil
	}
	sort.Strings(imageFiles)
	parts := strings.Fields(dirName)
	if len(parts) != 4 {
		return fmt.Errorf("unexpected directory name format: %s", dirName)
	}
	picsName := strings.Join(parts, "_")
	for i, file := range imageFiles {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%05d.jpg", picsName, i+1))
		if err := os.Rename(file, newPath); err != nil {
			return err
		}
	}
	return nil
}

// RenameDirectory renames a date-based directory and all images inside it
func (o *fileOrganiser) RenameDirectory(directory, newName string) error {
	// Clean the path to remove trailing slashes and normalize
	directory = filepath.Clean(directory)

	// Check if directory exists
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", directory)
	}

	// Convert to absolute path to ensure correct parent directory
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Extract base name and parse date
	baseName := filepath.Base(absDir)
	parts := strings.Fields(baseName)

	// Expect at least 4 parts: YYYY MM Month DD
	if len(parts) < 4 {
		return fmt.Errorf("directory name does not match expected format (YYYY MM Month DD [name]): %s", baseName)
	}

	// Build new directory name: date + new name
	dateParts := parts[:4]
	newDirName := strings.Join(dateParts, " ")
	if newName != "" {
		newDirName = newDirName + " " + newName
	}

	// Build full path for new directory
	parentDir := filepath.Dir(absDir)
	newDirPath := filepath.Join(parentDir, newDirName)

	logger.Debug("Rename paths", "original", directory, "absolute", absDir, "parent", parentDir, "new_name", newDirName, "new_path", newDirPath)

	// If the new path is the same as old, no directory rename needed
	if absDir == newDirPath {
		logger.Info("Directory name unchanged, updating images only")
	} else {
		// Check if target directory already exists
		if _, err := os.Stat(newDirPath); err == nil {
			return fmt.Errorf("target directory already exists: %s", newDirPath)
		}

		logger.Info("Renaming directory", "from", absDir, "to", newDirPath)
	}

	// Rename image files first (before moving directory)
	imageFiles := []string{}
	for _, pattern := range imageExtensions {
		files, err := filepath.Glob(filepath.Join(absDir, pattern))
		if err != nil {
			return fmt.Errorf("failed to find image files: %w", err)
		}
		imageFiles = append(imageFiles, files...)
	}

	if len(imageFiles) > 0 {
		sort.Strings(imageFiles)
		newBaseName := strings.ReplaceAll(newDirName, " ", "_")

		logger.Info("Renaming images", "count", len(imageFiles), "pattern", newBaseName)

		for i, file := range imageFiles {
			newFileName := fmt.Sprintf("%s_%05d.jpg", newBaseName, i+1)
			newFilePath := filepath.Join(absDir, newFileName)

			logger.Debug("Renaming image", "from", filepath.Base(file), "to", newFileName)

			if err := os.Rename(file, newFilePath); err != nil {
				return fmt.Errorf("failed to rename %s: %w", file, err)
			}
		}
	}

	// Rename videos in videos subdirectory if it exists
	videosDir := filepath.Join(absDir, "videos")
	if info, err := os.Stat(videosDir); err == nil && info.IsDir() {
		videoFiles := []string{}
		for _, pattern := range videoExtensions {
			files, err := filepath.Glob(filepath.Join(videosDir, pattern))
			if err != nil {
				return fmt.Errorf("failed to find video files: %w", err)
			}
			videoFiles = append(videoFiles, files...)
		}

		if len(videoFiles) > 0 {
			sort.Strings(videoFiles)
			newBaseName := strings.ReplaceAll(newDirName, " ", "_")
			logger.Info("Renaming videos", "count", len(videoFiles), "pattern", newBaseName)

			for i, file := range videoFiles {
				ext := filepath.Ext(file)
				newFileName := fmt.Sprintf("%s_%05d%s", newBaseName, i+1, ext)
				newFilePath := filepath.Join(videosDir, newFileName)
				logger.Debug("Renaming video", "from", filepath.Base(file), "to", newFileName)

				if err := os.Rename(file, newFilePath); err != nil {
					return fmt.Errorf("failed to rename %s: %w", file, err)
				}
			}
		}
	}

	// Rename the directory if needed
	if absDir != newDirPath {
		if err := os.Rename(absDir, newDirPath); err != nil {
			return fmt.Errorf("failed to rename directory: %w", err)
		}
		logger.Info("Directory renamed successfully", "new_path", newDirPath)
	}

	return nil
}
