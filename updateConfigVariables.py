#!/usr/bin/env python3
"""
updateConfigVariables.py

Description:
    Prompts the user for backendIP and BASE_DOMAIN (using previously saved values if available),
    then recursively searches through all .conf files under the conf.d directory and replaces 
    occurrences of ${backendIP} and ${BASE_DOMAIN} with the provided values. 
    A backup of any modified file is created before changes are saved.
"""

import os
import sys
import shutil
from datetime import datetime

LAST_VALUES_FILE = ".last_nginx.conf"
CONF_DIR = "conf.d"

def load_last_values():
    """Load saved values from LAST_VALUES_FILE, if it exists."""
    last_values = {}
    if os.path.isfile(LAST_VALUES_FILE):
        with open(LAST_VALUES_FILE, "r") as f:
            for line in f:
                line = line.strip()
                if not line or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                last_values[key.strip()] = value.strip().strip('"')
    return last_values

def prompt_input(var_name, prompt_message, default_val=None):
    """
    Prompt for input, displaying the default value if available.
    If the user inputs nothing and a default exists, return the default.
    Otherwise, reprompt until a non-empty value is entered.
    """
    while True:
        if default_val:
            user_input = input(f"{prompt_message} [{default_val}]: ").strip()
        else:
            user_input = input(f"{prompt_message}: ").strip()

        if not user_input and default_val:
            return default_val
        elif user_input:
            return user_input
        else:
            print(f"Error: {var_name} cannot be empty. Please enter a valid value.")

def save_last_values(values):
    """Save the provided values dictionary to LAST_VALUES_FILE."""
    with open(LAST_VALUES_FILE, "w") as f:
        for key, value in values.items():
            f.write(f'{key}="{value}"\n')

def backup_file(filepath):
    """If a file exists, back it up by copying it with a timestamp."""
    if os.path.isfile(filepath):
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        backup_path = os.path.join(os.path.dirname(filepath), f"{timestamp}_{os.path.basename(filepath)}.bak")
        shutil.copy2(filepath, backup_path)
        print(f"Backup of '{filepath}' created as '{backup_path}'.")

def update_file(filepath, backendIP, BASE_DOMAIN):
    """Replace the placeholders in a single file and update it if changes occur."""
    try:
        with open(filepath, "r") as f:
            content = f.read()
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        return

    new_content = content.replace("${backendIP}", backendIP).replace("${BASE_DOMAIN}", BASE_DOMAIN)

    if new_content != content:
        backup_file(filepath)
        try:
            with open(filepath, "w") as f:
                f.write(new_content)
            print(f"Updated {filepath}")
        except Exception as e:
            print(f"Error writing {filepath}: {e}")

def process_conf_directory(directory, backendIP, BASE_DOMAIN):
    """Recursively process all .conf files in the given directory."""
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith(".conf"):
                filepath = os.path.join(root, file)
                update_file(filepath, backendIP, BASE_DOMAIN)

def main():
    print("=== Recursive conf.d Variable Updater ===\n")

    # Load previous values if available
    last_values = load_last_values()

    # Prompt user for the backend IP and BASE_DOMAIN
    backendIP = prompt_input("backendIP", "Enter the backend IP address", last_values.get("backendIP"))
    BASE_DOMAIN = prompt_input("BASE_DOMAIN", "Enter the base domain for your services", last_values.get("BASE_DOMAIN"))

    # Save the values for future runs
    new_values = {"backendIP": backendIP, "BASE_DOMAIN": BASE_DOMAIN}
    save_last_values(new_values)

    # Check that the conf.d directory exists
    if not os.path.isdir(CONF_DIR):
        print(f"Error: Directory '{CONF_DIR}' not found in the current directory.")
        sys.exit(1)

    # Process all .conf files recursively in the conf.d directory
    process_conf_directory(CONF_DIR, backendIP, BASE_DOMAIN)

    print("\nDone updating configuration files in the conf.d directory.")

if __name__ == "__main__":
    main()
