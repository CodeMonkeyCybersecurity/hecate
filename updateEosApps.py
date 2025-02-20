#!/usr/bin/env python3
"""
updateEosApps.py

Description:
    Presents the user with a list of available EOS backend web apps (based on our supported apps).
    The user can select one or more (via comma-separated numbers).
    The script then recursively walks through the conf.d directory and deletes any .conf file 
    that is not associated with a selected app.
    
    Note: http.conf and stream.conf are always preserved because they are essential for the reverse proxy.

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
   12. Persephone       -> persephone.conf

"""

import os
import sys
import shutil
from datetime import datetime

LAST_VALUES_FILE = ".hecate.conf"
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

# Mapping of option number to (App Name, config filename)
APPS_SELECTION = {
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
    "12": ("Jellyfin", "jellyfin.conf"),
    "13": ("Persephone", "persephone.conf")
}

def display_options():
    print("Available EOS backend web apps:")
    for num in sorted(APPS_SELECTION, key=lambda x: int(x)):
        app_name, conf_file = APPS_SELECTION[num]
        print(f"  {num}. {app_name}  -> {conf_file}")

def get_user_selection(default_selection=None):
    """
    Prompts the user to enter a comma-separated list of option numbers.
    If a default is provided and the user enters nothing, the default is used.
    Returns a tuple of:
       - The set of allowed configuration filenames (user-selected)
       - The raw selection string (to save as the new default)
    """
    prompt_msg = "Enter the numbers (comma-separated) of the apps you want enabled (or type 'all' for all)"
    if default_selection:
        prompt_msg += f" [default: {default_selection}]"
    prompt_msg += ": "
    selection = input(prompt_msg).strip()
    if not selection and default_selection:
        selection = default_selection
    if selection.lower() == "all":
        # Return all app configuration files (user-selected ones)
        return set(APPS_SELECTION[num][1] for num in APPS_SELECTION), "all"
    chosen = set()
    valid = True
    for token in selection.split(","):
        token = token.strip()
        if token not in APPS_SELECTION:
            print(f"Invalid option: {token}")
            valid = False
            break
        chosen.add(APPS_SELECTION[token][1])
    if valid and chosen:
        return chosen, selection
    print("Please enter a valid comma-separated list of options.")
    return get_user_selection(default_selection)

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
    last_values = load_last_values()
    # Use the previously saved APPS value (if any) as the default selection.
    default_apps = last_values.get("APPS_SELECTION")
    display_options()
    allowed_files, selection_str = get_user_selection(default_apps)
    # Always add http.conf and stream.conf to the allowed list
    allowed_files.update({"http.conf", "stream.conf"})
    print("\nYou have selected the following configuration files to keep:")
    for f in allowed_files:
        # Check if the file is one of the essential files
        if f in {"http.conf", "stream.conf"}:
            print(f" - Essential file: {f}")
        else:
            for num, (app_name, conf_file) in APPS_SELECTION.items():
                if conf_file == f:
                    print(f" - {app_name} ({conf_file})")
    print("\nNow scanning the conf.d directory and removing files not in your selection...")
    remove_unwanted_conf_files(allowed_files)
    
    # Save the selection back into .hecate.conf
    last_values["APPS"] = selection_str
    save_last_values(last_values)

if __name__ == "__main__":
    main()
