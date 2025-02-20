#!/usr/bin/env python3
import os
import shutil
import sys

SRC_DIR = "conf.d"
BACKUP_DIR = "conf.d.bak"

if not os.path.isdir(SRC_DIR):
    print(f"Error: Source directory '{SRC_DIR}' does not exist.")
    sys.exit(1)

# Remove the backup directory if it exists
if os.path.exists(BACKUP_DIR):
    print(f"Removing existing backup directory '{BACKUP_DIR}'...")
    shutil.rmtree(BACKUP_DIR)

try:
    shutil.copytree(SRC_DIR, BACKUP_DIR)
    print(f"Backup complete: '{SRC_DIR}' has been backed up to '{BACKUP_DIR}'.")
except Exception as e:
    print(f"Error during backup: {e}")
    sys.exit(1)
