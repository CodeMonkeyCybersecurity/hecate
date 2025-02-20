#!/usr/bin/env python3
"""
updateEosApps.py

Description:
    Presents the user with a list of available EOS backend web apps (based on our supported apps).
    The user can select one or more (via comma-separated numbers).
    The script then recursively walks through the conf.d directory and deletes any .conf file 
    that is not associated with a selected app.

Available options and their corresponding configuration files:
    1. Static website   -> base.conf
    2. Wazuh            -> delphi.conf
    3. Mattermost       -> collaborate.conf
    4. Nextcloud        -> cloud.conf
    5. Mailcow          -> mailcow.conf
    6. Jenkins          -> jenkins.conf
    7. Grafana          -> observe.conf
    8. Umami            -> analytics.conf
    9. MinIO            -> s3.conf
   10. Wiki.js          -> wiki.conf
   11. ERPNext          -> erp.conf
   12. Jellyfin         -> jellyfin.conf
"""

import os
import sys

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
}

CONF_DIR = "conf.d"

def display_options():
    print("Available EOS backend web apps:")
    for num in sorted(APP_OPTIONS, key=lambda x: int(x)):
        app_name, conf_file = APP_OPTIONS[num]
        print(f"  {num}. {app_name}  -> {conf_file}")

def get_user_selection():
    while True:
        selection = input("Enter the numbers (comma-separated) of the apps you want enabled (or type 'all' for all): ").strip()
        if selection.lower() == "all":
            return set(APP_OPTIONS[num][1] for num in APP_OPTIONS)
        # Split by commas, remove whitespace, and validate each
        chosen = set()
        valid = True
        for token in selection.split(","):
            token = token.strip()
            if token not in APP_OPTIONS:
                print(f"Invalid option: {token}")
                valid = False
                break
            chosen.add(APP_OPTIONS[token][1])
        if valid and chosen:
            return chosen
        print("Please enter a valid comma-separated list of options.")

def remove_unwanted_conf_files(allowed_files):
    """
    Walks recursively through the CONF_DIR and removes any .conf file whose base name
    is not in the allowed_files set.
    """
    if not os.path.isdir(CONF_DIR):
        print(f"Error: Directory '{CONF_DIR}' not found.")
        sys.exit(1)

    removed_files = []
    for root, dirs, files in os.walk(CONF_DIR):
        for file in files:
            if file.endswith(".conf"):
                # If this config file is not in our allowed list, remove it.
                if file not in allowed_files:
                    full_path = os.path.join(root, file)
                    try:
                        os.remove(full_path)
                        removed_files.append(full_path)
                        print(f"Removed: {full_path}")
                    except Exception as e:
                        print(f"Error removing {full_path}: {e}")

    if not removed_files:
        print("No configuration files were removed.")
    else:
        print("\nCleanup complete. The following files were removed:")
        for f in removed_files:
            print(f" - {f}")

def main():
    print("=== EOS Backend Web Apps Selector ===\n")
    display_options()
    allowed_files = get_user_selection()
    print("\nYou have selected the following configuration files to keep:")
    for f in allowed_files:
        # Find the app name for display purposes
        for num, (app_name, conf_file) in APP_OPTIONS.items():
            if conf_file == f:
                print(f" - {app_name} ({conf_file})")
    print("\nNow scanning the conf.d directory and removing files not in your selection...")
    remove_unwanted_conf_files(allowed_files)

if __name__ == "__main__":
    main()
