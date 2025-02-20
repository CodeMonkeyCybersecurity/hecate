#!/usr/bin/env python3
import os
import shutil
import sys

BACKUP_DIR = "conf.d.bak"
DST_DIR = "conf.d"

if not os.path.isdir(BACKUP_DIR):
    print(f"Error: Backup directory '{BACKUP_DIR}' does not exist.")
    sys.exit(1)

# Remove the destination directory if it exists
if os.path.exists(DST_DIR):
    print(f"Removing current '{DST_DIR}' directory...")
    shutil.rmtree(DST_DIR)

try:
    shutil.copytree(BACKUP_DIR, DST_DIR)
    print(f"Restore complete: '{BACKUP_DIR}' has been restored to '{DST_DIR}'.")
except Exception as e:
    print(f"Error during restore: {e}")
    sys.exit(1)
