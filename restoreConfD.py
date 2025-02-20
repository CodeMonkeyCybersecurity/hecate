#!/usr/bin/env python3
import os
import shutil
import sys

# Define source (backup) and destination paths
BACKUP_CONF = "conf.d.bak"
BACKUP_CERTS = "certs.bak"
BACKUP_COMPOSE = "docker-compose.yml.bak"

DST_CONF = "conf.d"
DST_CERTS = "certs"
DST_COMPOSE = "docker-compose.yml"

def remove_if_exists(path):
    if os.path.exists(path):
        if os.path.isdir(path):
            print(f"Removing directory '{path}'...")
            shutil.rmtree(path)
        else:
            print(f"Removing file '{path}'...")
            os.remove(path)

# Restore conf.d directory
if not os.path.isdir(BACKUP_CONF):
    print(f"Error: Backup directory '{BACKUP_CONF}' does not exist.")
    sys.exit(1)
remove_if_exists(DST_CONF)
try:
    shutil.copytree(BACKUP_CONF, DST_CONF)
    print(f"Restore complete: '{BACKUP_CONF}' has been restored to '{DST_CONF}'.")
except Exception as e:
    print(f"Error during restore of {BACKUP_CONF}: {e}")
    sys.exit(1)

# Restore certs directory
if not os.path.isdir(BACKUP_CERTS):
    print(f"Error: Backup directory '{BACKUP_CERTS}' does not exist.")
    sys.exit(1)
remove_if_exists(DST_CERTS)
try:
    shutil.copytree(BACKUP_CERTS, DST_CERTS)
    print(f"Restore complete: '{BACKUP_CERTS}' has been restored to '{DST_CERTS}'.")
except Exception as e:
    print(f"Error during restore of {BACKUP_CERTS}: {e}")
    sys.exit(1)

# Restore docker-compose.yml file
if not os.path.isfile(BACKUP_COMPOSE):
    print(f"Error: Backup file '{BACKUP_COMPOSE}' does not exist.")
    sys.exit(1)
remove_if_exists(DST_COMPOSE)
try:
    shutil.copy2(BACKUP_COMPOSE, DST_COMPOSE)
    print(f"Restore complete: '{BACKUP_COMPOSE}' has been restored to '{DST_COMPOSE}'.")
except Exception as e:
    print(f"Error during restore of {BACKUP_COMPOSE}: {e}")
    sys.exit(1)
