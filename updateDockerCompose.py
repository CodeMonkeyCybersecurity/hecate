#!/usr/bin/env python3
"""
updateDockerCompose.py

This script reads the docker-compose.yml file and uncomments the appropriate port lines
based on the apps selected by the user.

Usage:
    # Command-line mode (non-interactive):
    python updateDockerCompose.py wazuh mailcow nextcloud

    # Interactive mode (if no command-line arguments are provided):
    python updateDockerCompose.py

Supported app keywords in the docker-compose file:
  - Wazuh
  - Mailcow
  - Nextcloud
"""

import sys
import re
import os
import shutil
from datetime import datetime

# File paths
LAST_VALUES_FILE = ".hecate.conf"
DOCKER_COMPOSE_FILE = "docker-compose.yml"

# Mapping of option number to (App Name, config filename)
APP_OPTIONS = {
    "1": ("Static website", "base.conf"),
    "2": ("Wazuh", "delphi.conf"),
    "3": ("Mattermost", "collaborate.conf"),
    "4": ("Nextcloud", "cloud.conf"),
    "5": ("Mailcow", "mailcow.conf"),
    "6": ("Jenkins", "jenkins.conf"),
    "7": ("Grafana", "observe.conf"),
    "8": ("Umami", "analytics.conf"),
    "9": ("MinIO", "s3.conf"),
    "10": ("Wiki.js", "wiki.conf"),
    "11": ("ERPNext", "erp.conf"),
    "12": ("Jellyfin", "jellyfin.conf")
    "13": ("Persephone", "persephone.conf")
}

# The docker-compose file only supports uncommenting lines for these apps.
# For each supported app, list the port markers you expect in the file.
SUPPORTED_APPS = {
    "wazuh": ["1515", "1514", "55000"],
    "mailcow": ["25", "587", "465", "110", "995", "143", "993"],
    "nextcloud": ["3478"]
}

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
    Reprompts until a non-empty value is entered.
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

def display_options():
    print("Available EOS backend web apps:")
    for num in sorted(APP_OPTIONS, key=lambda x: int(x)):
        app_name, conf_file = APP_OPTIONS[num]
        print(f"  {num}. {app_name}  -> {conf_file}")

def get_user_selection(default_selection=None):
    """
    Prompts the user to enter a comma-separated list of option numbers.
    If a default is provided and the user enters nothing, the default is used.
    Returns a tuple of:
       - A set of supported app keywords (from SUPPORTED_APPS) based on the selection.
       - The raw selection string (to save as the new default).
    """
    prompt_msg = "Enter the numbers (comma-separated) of the apps you want enabled (or type 'all' for all supported)"
    if default_selection:
        prompt_msg += f" [default: {default_selection}]"
    prompt_msg += ": "
    selection = input(prompt_msg).strip()
    if not selection and default_selection:
        selection = default_selection
    if selection.lower() == "all":
        return set(SUPPORTED_APPS.keys()), "all"
    
    chosen_keywords = set()
    valid = True
    for token in selection.split(","):
        token = token.strip()
        if token not in APP_OPTIONS:
            print(f"Invalid option: {token}")
            valid = False
            break
        # Map the app name to a supported keyword (if applicable)
        app_name, _ = APP_OPTIONS[token]
        key = app_name.lower()
        if key in SUPPORTED_APPS:
            chosen_keywords.add(key)
    if valid and chosen_keywords:
        return chosen_keywords, selection
    print("Please enter a valid comma-separated list of options corresponding to supported apps.")
    return get_user_selection(default_selection)

def update_compose_file(selected_apps):
    """
    Reads DOCKER_COMPOSE_FILE, and for each line that includes a marker corresponding to a
    supported app present in selected_apps, removes the leading '#' so that the line is uncommented.
    """
    try:
        with open(DOCKER_COMPOSE_FILE, "r") as f:
            lines = f.readlines()
    except FileNotFoundError:
        print(f"Error: {DOCKER_COMPOSE_FILE} not found.")
        sys.exit(1)

    new_lines = []
    # For each line, if it contains a marker for one of the selected apps, uncomment it.
    for line in lines:
        modified_line = line
        for app, markers in SUPPORTED_APPS.items():
            if app in selected_apps and any(marker in line for marker in markers):
                # Remove the leading '#' before '-' while preserving whitespace.
                modified_line = re.sub(r"^(\s*)#\s*(-)", r"\1\2", line)
                break  # if a match is found, no need to test other apps
        new_lines.append(modified_line)

    # Backup the original file
    backup_file(DOCKER_COMPOSE_FILE)

    with open(DOCKER_COMPOSE_FILE, "w") as f:
        f.writelines(new_lines)
    print(f"Updated {DOCKER_COMPOSE_FILE} for apps: {', '.join(selected_apps)}")

def main():
    last_values = load_last_values()
    # Determine if we are running in command-line mode or interactive mode
    if len(sys.argv) > 1:
        # Command-line mode: expect app keywords (e.g. wazuh, mailcow, nextcloud)
        selected_apps = {arg.lower() for arg in sys.argv[1:]} & set(SUPPORTED_APPS.keys())
        if not selected_apps:
            print("No supported apps found in the command-line arguments.")
            sys.exit(1)
        selection_str = ", ".join(selected_apps)
    else:
        # Interactive mode: show options and prompt the user.
        display_options()
        default_selection = last_values.get("APPS_SELECTION")
        selected_apps, selection_str = get_user_selection(default_selection)
        # Save this selection as the new default.
        last_values["APPS_SELECTION"] = selection_str
        save_last_values(last_values)

    update_compose_file(selected_apps)

if __name__ == "__main__":
    main()
