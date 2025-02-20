#!/usr/bin/env python3
"""
generateNginxConf.py

Description:
    Prompts the user for backendIP and BASE_DOMAIN, then replaces placeholders in
    nginx.conf.template with these values. It also remembers the last-used values in
    a file named .last_nginx.conf.
"""

import os
import sys
import shutil
from datetime import datetime

LAST_VALUES_FILE = ".last_nginx.conf"
TEMPLATE_FILE = "nginx.conf.template"
OUTPUT_FILE = "nginx.conf"


def load_last_values():
    """Load saved values from LAST_VALUES_FILE, if it exists."""
    last_values = {}
    if os.path.isfile(LAST_VALUES_FILE):
        with open(LAST_VALUES_FILE, "r") as f:
            for line in f:
                # Expecting lines in the format: key="value"
                line = line.strip()
                if not line or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                # Remove any surrounding quotes
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
        backup_path = f"{timestamp}_{os.path.basename(filepath)}.bak"
        shutil.copy2(filepath, backup_path)
        print(f"Backup of existing '{filepath}' created as '{backup_path}'.")


def generate_nginx_conf(backendIP, BASE_DOMAIN):
    """Replace placeholders in the template file and write to OUTPUT_FILE."""
    if not os.path.isfile(TEMPLATE_FILE):
        print(f"Error: Template file '{TEMPLATE_FILE}' not found in the current directory.")
        sys.exit(1)

    # Read the template
    with open(TEMPLATE_FILE, "r") as f:
        template_content = f.read()

    # Replace placeholders
    # We assume placeholders in the template look like ${backendIP} and ${BASE_DOMAIN}
    output_content = template_content.replace("${backendIP}", backendIP).replace("${BASE_DOMAIN}", BASE_DOMAIN)

    # Backup existing output file if it exists
    backup_file(OUTPUT_FILE)

    # Write the new configuration file
    with open(OUTPUT_FILE, "w") as f:
        f.write(output_content)

    print(f"'{OUTPUT_FILE}' generated successfully using the provided values.")


def main():
    print("=== NGINX Configuration Generator ===\n")

    # Load previous values, if any
    last_values = load_last_values()

    # Prompt for backendIP and BASE_DOMAIN, using defaults if available
    backendIP = prompt_input("backendIP", "Enter the backend IP address", last_values.get("backendIP"))
    BASE_DOMAIN = prompt_input("BASE_DOMAIN", "Enter the base domain for your services", last_values.get("BASE_DOMAIN"))

    # Save the values for future runs
    new_values = {"backendIP": backendIP, "BASE_DOMAIN": BASE_DOMAIN}
    save_last_values(new_values)

    # Generate the new nginx.conf file
    generate_nginx_conf(backendIP, BASE_DOMAIN)


if __name__ == "__main__":
    main()
