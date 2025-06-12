package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func CheckIfFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// The path does not exist
		return false, nil
	}
	if err != nil {
		// Some other error occurred (e.g., permissions, invalid path)
		return false, fmt.Errorf("error checking path %s: %v", path, err)
	}
	// The path exists, now check if it's a file
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("path %s exists but is not a regular file", path)
	}
	return true, nil // It exists and is a regular file
}

func CheckIfDirExists(path string) (bool, error) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		// The path does not exist
		return false, nil
	}
	if err != nil {
		// Some other error occurred (e.g., permissions, invalid path)
		return false, fmt.Errorf("error checking path %s: %v", path, err)
	}

	// The path exists, now check if it's a directory
	if !info.IsDir() {
		return false, fmt.Errorf("path %s exists but is not a directory", path)
	}

	return true, nil // It exists and is a directory
}

func CreateDirIfNotExists(path string) error {
	// Check if the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the directory if it does not exist
		err = os.MkdirAll(path, 0755) // Ensure the directory is created with appropriate permissions
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}
	return nil
}

func CreateDirFileIfNotExists(path string) error {
	// Check if the directory exists
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		// Create the directory if it does not exist
		err = os.MkdirAll(filepath.Dir(path), 0755) // Ensure the directory is created with appropriate permissions
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}

	// Create a file in the directory if it does not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", path, err)
		}
		defer file.Close()
	}

	return nil
}

func ReadFile(ctx context.Context, filePath string) (string, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Ensure the file exists
	if exists, err := CheckIfFileExists(filePath); err != nil {
		return "", fmt.Errorf("error checking if file exists %s: %v", filePath, err)
	} else if !exists {
		return "", fmt.Errorf("file %s does not exist", filePath)
	}

	// Read the content of the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	logger.Infof("Read content from file %s\n", filePath)
	return string(data), nil
}

func AppendToFile(ctx context.Context, filePath, content string, mode os.FileMode) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Ensure the file exists
	if err := CreateDirFileIfNotExists(filePath); err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}

	// Open the file for appending
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("failed to open file %s for appending: %v", filePath, err)
	}
	defer file.Close()

	// Write the content to the file
	if _, err := file.WriteString(content + "\n"); err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}

	// Ensure the file is closed properly
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file %s: %v", filePath, err)
	}

	logger.Infof("Appended content to file %s\n", filePath)
	return nil
}

func WriteToFile(ctx context.Context, filePath, content string, mode os.FileMode) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Ensure the directory exists
	if err := CreateDirFileIfNotExists(filePath); err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(content), mode); err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}

	logger.Infof("Wrote content to file %s\n", filePath)
	return nil
}

func RenameFile(ctx context.Context, oldPath, newPath string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Ensure the old file exists
	if exists, err := CheckIfFileExists(oldPath); err != nil {
		return fmt.Errorf("error checking if file exists %s: %v", oldPath, err)
	} else if !exists {
		return fmt.Errorf("file %s does not exist", oldPath)
	}

	// Rename the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename file from %s to %s: %v", oldPath, newPath, err)
	}

	logger.Infof("Renamed file from %s to %s\n", oldPath, newPath)
	return nil
}
