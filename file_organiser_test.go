package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper functions

func createDirs(t *testing.T, tmpDir string) (string, string) {
	t.Helper()
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	return sourceDir, targetDir
}

func createFileWithDate(t *testing.T, dir, filename string, modTime time.Time) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", filename, err)
	}
	if err := os.Chtimes(filePath, modTime, modTime); err != nil {
		t.Fatalf("Failed to set file times: %v", err)
	}
	return filePath
}

func createDateDir(t *testing.T, parentDir, dateName string) string {
	t.Helper()
	dirPath := filepath.Join(parentDir, dateName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create date directory: %v", err)
	}
	return dirPath
}

func createFile(t *testing.T, dir, filename string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", filename, err)
	}
	return filePath
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Expected file to exist at %s", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected file to not exist at %s", path)
	}
}

func TestFileOrganiser_OrganiseByDate(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir, targetDir := createDirs(t, tmpDir)

	// Create test files with specific modification times
	testDate := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
	file1 := createFileWithDate(t, sourceDir, "image1.jpg", testDate)
	file2 := createFileWithDate(t, sourceDir, "image2.jpeg", testDate)

	// Organise files by date
	organiser := NewFileOrganiser()
	err := organiser.OrganiseByDate(sourceDir, targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that files were moved to date-based directory
	expectedDir := filepath.Join(targetDir, "2023 06 June 15")
	assertFileExists(t, expectedDir)
	assertFileExists(t, filepath.Join(expectedDir, "image1.jpg"))
	assertFileExists(t, filepath.Join(expectedDir, "image2.jpeg"))

	// Check files were removed from source
	assertFileNotExists(t, file1)
	assertFileNotExists(t, file2)
}

func TestFileOrganiser_OrganiseByDate_MultipleDates(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir, targetDir := createDirs(t, tmpDir)

	// Create files with different dates
	date1 := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
	date2 := time.Date(2023, 7, 20, 14, 0, 0, 0, time.UTC)

	createFileWithDate(t, sourceDir, "june.jpg", date1)
	createFileWithDate(t, sourceDir, "july.jpg", date2)

	organiser := NewFileOrganiser()
	err := organiser.OrganiseByDate(sourceDir, targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check both date directories were created
	assertFileExists(t, filepath.Join(targetDir, "2023 06 June 15", "june.jpg"))
	assertFileExists(t, filepath.Join(targetDir, "2023 07 July 20", "july.jpg"))
}

func TestFileOrganiser_OrganiseByDate_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir, targetDir := createDirs(t, tmpDir)

	// Create a subdirectory in source
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file in source root
	testDate := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
	createFileWithDate(t, sourceDir, "image1.jpg", testDate)

	organiser := NewFileOrganiser()
	err := organiser.OrganiseByDate(sourceDir, targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check file was moved
	assertFileExists(t, filepath.Join(targetDir, "2023 06 June 15", "image1.jpg"))

	// Check subdirectory still exists in source
	assertFileExists(t, subDir)
}

func TestFileOrganiser_OrganiseByDate_NonexistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseByDate("/nonexistent/source", targetDir)

	if err == nil {
		t.Error("Expected error for nonexistent source directory")
	}
}

func TestFileOrganiser_OrganiseVideosAndRenameImages(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)
	dateDir := createDateDir(t, targetDir, "2023 06 June 15")

	// Create test images and videos
	createFile(t, dateDir, "img1.jpg")
	createFile(t, dateDir, "img2.heic")
	createFile(t, dateDir, "vid1.mov")
	createFile(t, dateDir, "vid2.MOV")

	// Organise videos and rename images
	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check images were renamed
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00001.jpg"))
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00002.heic"))

	// Check videos were moved to subdirectory and renamed
	videosDir := filepath.Join(dateDir, "videos")
	assertFileExists(t, videosDir)
	assertFileExists(t, filepath.Join(videosDir, "2023_06_June_15_00001.mov"))
	assertFileExists(t, filepath.Join(videosDir, "2023_06_June_15_00002.MOV"))

	// Check videos were removed from root
	assertFileNotExists(t, filepath.Join(dateDir, "vid1.mov"))
	assertFileNotExists(t, filepath.Join(dateDir, "vid2.MOV"))
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_OnlyImages(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)
	dateDir := createDateDir(t, targetDir, "2023 06 June 15")

	// Create only images
	createFile(t, dateDir, "img1.jpg")
	createFile(t, dateDir, "img2.jpeg")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check images were renamed
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00001.jpg"))
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00002.jpeg"))

	// Check videos subdirectory was not created
	assertFileNotExists(t, filepath.Join(dateDir, "videos"))
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_OnlyVideos(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)
	dateDir := createDateDir(t, targetDir, "2023 06 June 15")

	// Create only videos
	createFile(t, dateDir, "vid1.mov")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check videos subdirectory was created
	videosDir := filepath.Join(dateDir, "videos")
	assertFileExists(t, videosDir)
	assertFileExists(t, filepath.Join(videosDir, "2023_06_June_15_00001.mov"))
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)

	// Create an empty date-based directory
	dateDir := createDateDir(t, targetDir, "2023 06 June 15")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Directory should still exist
	assertFileExists(t, dateDir)
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_InvalidDirectoryFormat(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)

	// Create a directory with invalid format
	invalidDir := createDateDir(t, targetDir, "invalid format")

	// Create a file in the invalid directory
	createFile(t, invalidDir, "img.jpg")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err == nil {
		t.Error("Expected error for invalid directory format")
	}
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_SkipsFiles(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)
	dateDir := createDateDir(t, targetDir, "2023 06 June 15")

	// Create a file directly in target (should be skipped)
	createFile(t, targetDir, "random.txt")

	// Create an image in the date directory
	createFile(t, dateDir, "img1.jpg")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check image was renamed
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00001.jpg"))

	// Check random file still exists and was not processed
	assertFileExists(t, filepath.Join(targetDir, "random.txt"))
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_NonexistentTarget(t *testing.T) {
	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages("/nonexistent/target")

	if err == nil {
		t.Error("Expected error for nonexistent target directory")
	}
}

func TestFileOrganiser_OrganiseVideosAndRenameImages_PreservesExtensionCase(t *testing.T) {
	tmpDir := t.TempDir()
	_, targetDir := createDirs(t, tmpDir)
	dateDir := createDateDir(t, targetDir, "2023 06 June 15")

	// Create images with different extension cases
	createFile(t, dateDir, "img1.JPG")
	createFile(t, dateDir, "img2.HEIC")

	// Create video with uppercase extension
	createFile(t, dateDir, "vid1.MOV")

	organiser := NewFileOrganiser()
	err := organiser.OrganiseVideosAndRenameImages(targetDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check images have lowercase extensions
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00001.jpg"))
	assertFileExists(t, filepath.Join(dateDir, "2023_06_June_15_00002.heic"))

	// Check video preserves original extension case
	videosDir := filepath.Join(dateDir, "videos")
	assertFileExists(t, filepath.Join(videosDir, "2023_06_June_15_00001.MOV"))
}

func TestFileOrganiser_GetFileCount(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	createFile(t, tmpDir, "file1.txt")
	createFile(t, tmpDir, "file2.jpg")
	createFile(t, tmpDir, "file3.mov")

	// Create subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	createFile(t, subDir, "file4.txt")
	createFile(t, subDir, "file5.jpg")

	organiser := NewFileOrganiser()
	count, err := organiser.GetFileCount(tmpDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should count all 5 files
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestFileOrganiser_GetFileCount_SkipsDotFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create regular files
	createFile(t, tmpDir, "file1.txt")
	createFile(t, tmpDir, "file2.jpg")

	// Create dot files (should be skipped)
	createFile(t, tmpDir, ".hidden")
	createFile(t, tmpDir, ".DS_Store")

	// Create dot directory with files (should be skipped)
	dotDir := filepath.Join(tmpDir, ".dotdir")
	if err := os.MkdirAll(dotDir, 0755); err != nil {
		t.Fatalf("Failed to create dot directory: %v", err)
	}
	createFile(t, dotDir, "file3.txt")

	organiser := NewFileOrganiser()
	count, err := organiser.GetFileCount(tmpDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should count only 2 regular files, skipping dot files and dot directories
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestFileOrganiser_GetFileCount_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	organiser := NewFileOrganiser()
	count, err := organiser.GetFileCount(tmpDir)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should count 0 files in empty directory
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestFileOrganiser_GetFileCount_NonexistentDirectory(t *testing.T) {
	organiser := NewFileOrganiser()
	_, err := organiser.GetFileCount("/nonexistent/directory")

	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}
}
