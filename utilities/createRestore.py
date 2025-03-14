package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// removeIfExists removes the file or directory at the given path if it exists.
func removeIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			fmt.Printf("Removing directory '%s'...\n", path)
			return os.RemoveAll(path)
		} else {
			fmt.Printf("Removing file '%s'...\n", path)
			return os.Remove(path)
		}
	}
	return nil
}

// copyFile copies a file from src to dst.
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

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// Preserve file permissions.
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

// copyDir recursively copies a directory tree from src to dst.
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

	// Create destination directory with the same permissions.
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	// Define backup (source) and destination paths.
	const (
		BACKUP_CONF    = "conf.d.bak"
		BACKUP_CERTS   = "certs.bak"
		BACKUP_COMPOSE = "docker-compose.yml.bak"

		DST_CONF    = "conf.d"
		DST_CERTS   = "certs"
		DST_COMPOSE = "docker-compose.yml"
	)

	// Restore conf.d directory.
	info, err := os.Stat(BACKUP_CONF)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Backup directory '%s' does not exist.\n", BACKUP_CONF)
		os.Exit(1)
	}
	if err := removeIfExists(DST_CONF); err != nil {
		fmt.Printf("Error removing %s: %v\n", DST_CONF, err)
		os.Exit(1)
	}
	if err := copyDir(BACKUP_CONF, DST_CONF); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", BACKUP_CONF, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", BACKUP_CONF, DST_CONF)

	// Restore certs directory.
	info, err = os.Stat(BACKUP_CERTS)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: Backup directory '%s' does not exist.\n", BACKUP_CERTS)
		os.Exit(1)
	}
	if err := removeIfExists(DST_CERTS); err != nil {
		fmt.Printf("Error removing %s: %v\n", DST_CERTS, err)
		os.Exit(1)
	}
	if err := copyDir(BACKUP_CERTS, DST_CERTS); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", BACKUP_CERTS, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", BACKUP_CERTS, DST_CERTS)

	// Restore docker-compose.yml file.
	info, err = os.Stat(BACKUP_COMPOSE)
	if err != nil || info.IsDir() {
		fmt.Printf("Error: Backup file '%s' does not exist.\n", BACKUP_COMPOSE)
		os.Exit(1)
	}
	if err := removeIfExists(DST_COMPOSE); err != nil {
		fmt.Printf("Error removing %s: %v\n", DST_COMPOSE, err)
		os.Exit(1)
	}
	if err := copyFile(BACKUP_COMPOSE, DST_COMPOSE); err != nil {
		fmt.Printf("Error during restore of %s: %v\n", BACKUP_COMPOSE, err)
		os.Exit(1)
	}
	fmt.Printf("Restore complete: '%s' has been restored to '%s'.\n", BACKUP_COMPOSE, DST_COMPOSE)
}
