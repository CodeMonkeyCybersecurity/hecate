package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// removeIfExists deletes a file or directory if it exists.
func removeIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			fmt.Printf("Removing existing directory '%s'...\n", path)
			return os.RemoveAll(path)
		}
		fmt.Printf("Removing existing file '%s'...\n", path)
		return os.Remove(path)
	}
	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	// Ensure the copied file has the same permissions.
	if info, err := os.Stat(src); err == nil {
		return os.Chmod(dst, info.Mode())
	}
	return nil
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source %s is not a directory", src)
	}

	// Create destination directory.
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}

		if entryInfo.IsDir() {
			if err = copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	const (
		SRC_CONF    = "conf.d"
		SRC_CERTS   = "certs"
		SRC_COMPOSE = "docker-compose.yml"

		BACKUP_CONF    = "conf.d.bak"
		BACKUP_CERTS   = "certs.bak"
		BACKUP_COMPOSE = "docker-compose.yml.bak"
	)

	// Backup conf.d directory
	confInfo, err := os.Stat(SRC_CONF)
	if err != nil || !confInfo.IsDir() {
		fmt.Printf("Error: Source directory '%s' does not exist.\n", SRC_CONF)
		os.Exit(1)
	}
	if err := removeIfExists(BACKUP_CONF); err != nil {
		fmt.Printf("Error removing backup directory '%s': %v\n", BACKUP_CONF, err)
		os.Exit(1)
	}
	if err := copyDir(SRC_CONF, BACKUP_CONF); err != nil {
		fmt.Printf("Error during backup of %s: %v\n", SRC_CONF, err)
		os.Exit(1)
	}
	fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_CONF, BACKUP_CONF)

	// Backup certs directory
	certsInfo, err := os.Stat(SRC_CERTS)
	if err != nil || !certsInfo.IsDir() {
		fmt.Printf("Error: Source directory '%s' does not exist.\n", SRC_CERTS)
		os.Exit(1)
	}
	if err := removeIfExists(BACKUP_CERTS); err != nil {
		fmt.Printf("Error removing backup directory '%s': %v\n", BACKUP_CERTS, err)
		os.Exit(1)
	}
	if err := copyDir(SRC_CERTS, BACKUP_CERTS); err != nil {
		fmt.Printf("Error during backup of %s: %v\n", SRC_CERTS, err)
		os.Exit(1)
	}
	fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_CERTS, BACKUP_CERTS)

	// Backup docker-compose.yml file
	composeInfo, err := os.Stat(SRC_COMPOSE)
	if err != nil || composeInfo.IsDir() {
		fmt.Printf("Error: Source file '%s' does not exist.\n", SRC_COMPOSE)
		os.Exit(1)
	}
	if err := removeIfExists(BACKUP_COMPOSE); err != nil {
		fmt.Printf("Error removing backup file '%s': %v\n", BACKUP_COMPOSE, err)
		os.Exit(1)
	}
	if err := copyFile(SRC_COMPOSE, BACKUP_COMPOSE); err != nil {
		fmt.Printf("Error during backup of %s: %v\n", SRC_COMPOSE, err)
		os.Exit(1)
	}
	fmt.Printf("Backup complete: '%s' has been backed up to '%s'.\n", SRC_COMPOSE, BACKUP_COMPOSE)
}
