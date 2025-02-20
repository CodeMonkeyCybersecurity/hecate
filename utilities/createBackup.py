#!/usr/bin/env python3
import os
import shutil
import sys

# Define source and backup paths
SRC_CONF = "conf.d"
SRC_CERTS = "certs"
SRC_COMPOSE = "docker-compose.yml"

BACKUP_CONF = "conf.d.bak"
BACKUP_CERTS = "certs.bak"
BACKUP_COMPOSE = "docker-compose.yml.bak"

def remove_if_exists(path):
    if os.path.exists(path):
        if os.path.isdir(path):
            print(f"Removing existing directory '{path}'...")
            shutil.rmtree(path)
        else:
            print(f"Removing existing file '{path}'...")
            os.remove(path)

# Backup conf.d directory
if not os.path.isdir(SRC_CONF):
    print(f"Error: Source directory '{SRC_CONF}' does not exist.")
    sys.exit(1)
remove_if_exists(BACKUP_CONF)
try:
    shutil.copytree(SRC_CONF, BACKUP_CONF)
    print(f"Backup complete: '{SRC_CONF}' has been backed up to '{BACKUP_CONF}'.")
except Exception as e:
    print(f"Error during backup of {SRC_CONF}: {e}")
    sys.exit(1)

# Backup certs directory
if not os.path.isdir(SRC_CERTS):
    print(f"Error: Source directory '{SRC_CERTS}' does not exist.")
    sys.exit(1)
remove_if_exists(BACKUP_CERTS)
try:
    shutil.copytree(SRC_CERTS, BACKUP_CERTS)
    print(f"Backup complete: '{SRC_CERTS}' has been backed up to '{BACKUP_CERTS}'.")
except Exception as e:
    print(f"Error during backup of {SRC_CERTS}: {e}")
    sys.exit(1)

# Backup docker-compose.yml file
if not os.path.isfile(SRC_COMPOSE):
    print(f"Error: Source file '{SRC_COMPOSE}' does not exist.")
    sys.exit(1)
remove_if_exists(BACKUP_COMPOSE)
try:
    shutil.copy2(SRC_COMPOSE, BACKUP_COMPOSE)
    print(f"Backup complete: '{SRC_COMPOSE}' has been backed up to '{BACKUP_COMPOSE}'.")
except Exception as e:
    print(f"Error during backup of {SRC_COMPOSE}: {e}")
    sys.exit(1)
