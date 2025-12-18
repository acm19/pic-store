package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		parse()
	case "rename":
		rename()
	case "backup":
		backup()
	case "restore":
		restore()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  %s parse SOURCE_DIR TARGET_DIR         Process and organise media files
  %s rename DIRECTORY NAME               Rename a date-based directory and its images
  %s backup SOURCE_DIR BUCKET            Backup directories to S3
  %s restore BUCKET TARGET_DIR [OPTIONS] Restore directories from S3

Restore options (environment variables):
  FROM=YYYY or MM/YYYY    Lower bound (optional)
  TO=YYYY or MM/YYYY      Upper bound (optional)
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func parse() {
	if len(os.Args) != 4 {
		logger.Error("Invalid arguments", "usage", os.Args[0]+" parse SOURCE_DIR TARGET_DIR")
		os.Exit(1)
	}

	sourceDir := os.Args[2]
	targetDir := os.Args[3]

	parser := NewMediaParser()
	if err := parser.ValidateDirectories(sourceDir, targetDir); err != nil {
		logger.Error("Directory validation failed", "error", err)
		os.Exit(1)
	}

	opts := DefaultParseOptions()
	if rate := os.Getenv("RATE"); rate != "" {
		quality, err := strconv.Atoi(rate)
		if err != nil {
			logger.Error("Invalid RATE value", "rate", rate, "error", err)
			os.Exit(1)
		}
		opts.JPEGQuality = quality
	}

	sourceCount, err := parser.GetFileCount(sourceDir)
	if err != nil {
		logger.Error("Error counting source files", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting media parsing", "source", sourceDir, "target", targetDir)
	if err := parser.Parse(sourceDir, targetDir, opts); err != nil {
		logger.Error("Parse failed", "error", err)
		os.Exit(1)
	}

	targetCount, err := parser.GetFileCount(targetDir)
	if err != nil {
		logger.Error("Error counting target files", "error", err)
		os.Exit(1)
	}

	if sourceCount != targetCount {
		logger.Error("File count mismatch", "source_files", sourceCount, "target_files", targetCount, "difference", targetCount-sourceCount)
		os.Exit(1)
	}

	logger.Info("Processing completed successfully", "files_processed", sourceCount, "verification", "source and target file counts match")
}

func rename() {
	if len(os.Args) != 4 {
		logger.Error("Invalid arguments", "usage", os.Args[0]+" rename DIRECTORY NAME")
		os.Exit(1)
	}

	directory := os.Args[2]
	newName := os.Args[3]

	organiser := NewFileOrganiser()
	if err := organiser.RenameDirectory(directory, newName); err != nil {
		logger.Error("Rename failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Rename completed successfully")
}

func backup() {
	if len(os.Args) != 4 {
		logger.Error("Invalid arguments", "usage", os.Args[0]+" backup SOURCE_DIR BUCKET")
		os.Exit(1)
	}

	sourceDir := os.Args[2]
	bucket := os.Args[3]

	// Get max concurrent workers from environment variable, default to 5
	maxConcurrent := 5
	if concurrency := os.Getenv("MAX_CONCURRENT"); concurrency != "" {
		parsedConcurrency, err := strconv.Atoi(concurrency)
		if err != nil {
			logger.Error("Invalid MAX_CONCURRENT value", "value", concurrency, "error", err)
			os.Exit(1)
		}
		maxConcurrent = parsedConcurrency
	}

	// Validate source directory exists
	if info, err := os.Stat(sourceDir); err != nil {
		logger.Error("Source directory does not exist", "directory", sourceDir, "error", err)
		os.Exit(1)
	} else if !info.IsDir() {
		logger.Error("Source path is not a directory", "path", sourceDir)
		os.Exit(1)
	}

	// Create S3 backup instance
	ctx := context.Background()
	s3Backup, err := NewS3Backup(ctx)
	if err != nil {
		logger.Error("Failed to initialize S3 client", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting backup", "source", sourceDir, "bucket", bucket, "max_concurrent", maxConcurrent)
	if err := s3Backup.BackupDirectories(sourceDir, bucket, maxConcurrent); err != nil {
		logger.Error("Backup failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Backup completed successfully")
}

func restore() {
	if len(os.Args) != 4 {
		logger.Error("Invalid arguments", "usage", os.Args[0]+" restore BUCKET TARGET_DIR")
		os.Exit(1)
	}

	bucket := os.Args[2]
	targetDir := os.Args[3]

	// Get max concurrent workers from environment variable, default to 5
	maxConcurrent := 5
	if concurrency := os.Getenv("MAX_CONCURRENT"); concurrency != "" {
		parsedConcurrency, err := strconv.Atoi(concurrency)
		if err != nil {
			logger.Error("Invalid MAX_CONCURRENT value", "value", concurrency, "error", err)
			os.Exit(1)
		}
		maxConcurrent = parsedConcurrency
	}

	// Parse filter from environment variables
	var filter RestoreFilter

	if fromStr := os.Getenv("FROM"); fromStr != "" {
		year, month, err := parseYearMonth(fromStr)
		if err != nil {
			logger.Error("Invalid FROM value (expected YYYY or MM/YYYY)", "value", fromStr, "error", err)
			os.Exit(1)
		}
		filter.FromYear = year
		filter.FromMonth = month
	}

	if toStr := os.Getenv("TO"); toStr != "" {
		year, month, err := parseYearMonth(toStr)
		if err != nil {
			logger.Error("Invalid TO value (expected YYYY or MM/YYYY)", "value", toStr, "error", err)
			os.Exit(1)
		}
		filter.ToYear = year
		filter.ToMonth = month
	}

	// Validate target directory exists
	if info, err := os.Stat(targetDir); err != nil {
		logger.Error("Target directory does not exist", "directory", targetDir, "error", err)
		os.Exit(1)
	} else if !info.IsDir() {
		logger.Error("Target path is not a directory", "path", targetDir)
		os.Exit(1)
	}

	// Create S3 backup instance
	ctx := context.Background()
	s3Backup, err := NewS3Backup(ctx)
	if err != nil {
		logger.Error("Failed to initialize S3 client", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting restore", "bucket", bucket, "target", targetDir, "max_concurrent", maxConcurrent, "filter", filter)
	if err := s3Backup.RestoreDirectories(bucket, targetDir, filter, maxConcurrent); err != nil {
		logger.Error("Restore failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Restore completed successfully")
}

// parseYearMonth parses a date string in format "YYYY" or "MM/YYYY"
// Returns (year, month, error). Month is 0 if not specified.
func parseYearMonth(s string) (int, int, error) {
	parts := strings.Split(s, "/")

	if len(parts) == 1 {
		// Format: YYYY
		year, err := strconv.Atoi(parts[0])
		if err != nil || year < 1000 || year > 9999 {
			return 0, 0, fmt.Errorf("invalid year: %s", parts[0])
		}
		return year, 0, nil
	} else if len(parts) == 2 {
		// Format: MM/YYYY
		month, err := strconv.Atoi(parts[0])
		if err != nil || month < 1 || month > 12 {
			return 0, 0, fmt.Errorf("invalid month (must be 1-12): %s", parts[0])
		}
		year, err := strconv.Atoi(parts[1])
		if err != nil || year < 1000 || year > 9999 {
			return 0, 0, fmt.Errorf("invalid year: %s", parts[1])
		}
		return year, month, nil
	}

	return 0, 0, fmt.Errorf("invalid format (expected YYYY or MM/YYYY): %s", s)
}
