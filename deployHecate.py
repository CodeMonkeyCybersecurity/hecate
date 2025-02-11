#!/usr/bin/env python3
"""
deploy.py

A simple interactive script to deploy the reverse proxy docker-compose stack.
It asks the user which applications to deploy (e.g. static site, Wazuh, Mailcow),
the backend IP, and the base domain. If Wazuh (option 2) or Mailcow (option 3) is
selected, the final configuration will include both stream and HTTP blocks.

It also checks that the required TLS certificate files are present before deploying.
"""

import os
import subprocess
import sys

# Define the mapping of application numbers to names.
APP_OPTIONS = {
    1: "Static HTTP website",
    2: "Wazuh",
    3: "Mailcow",
    4: "Umami",
    5: "Mattermost",
    6: "Nextcloud",
    7: "ERPNext",
    8: "Jellyfin",
    9: "Grafana",
    10: "Minio",
    11: "Jenkins"
}


def get_applications():
    """Prompt the user to select which applications to deploy."""
    print("Select which applications to deploy behind the reverse proxy:")
    for num, name in APP_OPTIONS.items():
        print(f"  {num}) {name}")
    choices = input("Enter comma-separated numbers (e.g., 1,3): ")
    try:
        selected = [int(x.strip()) for x in choices.split(",") if x.strip().isdigit()]
    except ValueError:
        print("Invalid input. Please enter numbers separated by commas.")
        sys.exit(1)
    # Validate choices
    for choice in selected:
        if choice not in APP_OPTIONS:
            print(f"Invalid option: {choice}")
            sys.exit(1)
    return selected


def check_certificates(base_domain, selected_apps):
    """
    Check for required TLS certificate files.
    
    For the base domain, we expect:
      ./certs/<base_domain>.fullchain.pem
      ./certs/<base_domain>.privkey.pem

    For Mailcow (option 3), we might require additional certs:
      ./certs/mail.fullchain.pem and ./certs/mail.privkey.pem
    """
    missing_files = []

    # Base domain certificates
    base_fullchain = f"./certs/{base_domain}.fullchain.pem"
    base_privkey = f"./certs/{base_domain}.privkey.pem"
    if not os.path.exists(base_fullchain):
        missing_files.append(base_fullchain)
    if not os.path.exists(base_privkey):
        missing_files.append(base_privkey)

    # For Mailcow, check for mail certificates.
    if 3 in selected_apps:
        mail_fullchain = "./certs/mail.fullchain.pem"
        mail_privkey = "./certs/mail.privkey.pem"
        if not os.path.exists(mail_fullchain):
            missing_files.append(mail_fullchain)
        if not os.path.exists(mail_privkey):
            missing_files.append(mail_privkey)

    if missing_files:
        print("\nWARNING: The following TLS certificate files are missing:")
        for f in missing_files:
            print(f"  - {f}")
        print("Please make sure all required certificate files are available before deploying.")
    else:
        print("All required certificate files are present.\n")


def write_env_file(backend_ip, base_domain, include_stream):
    """
    Write a .env file with the provided values.
    The docker-compose file should read variables (e.g. BACKEND_IP, BASE_DOMAIN,
    and INCLUDE_STREAM) from this file.
    """
    env_content = (
        f"BACKEND_IP={backend_ip}\n"
        f"BASE_DOMAIN={base_domain}\n"
        f"INCLUDE_STREAM={'yes' if include_stream else 'no'}\n"
    )
    try:
        with open(".env", "w") as env_file:
            env_file.write(env_content)
        print("Environment file (.env) written successfully.\n")
    except Exception as e:
        print(f"Error writing .env file: {e}")
        sys.exit(1)


def deploy_stack():
    """Bring the docker-compose stack down and then up in detached mode."""
    print("Tearing down any existing containers...")
    subprocess.run("docker-compose down", shell=True, check=True)
    print("Starting up the reverse proxy stack in detached mode...")
    subprocess.run("docker-compose up -d", shell=True, check=True)
    print("Deployment successful!\n")


def main():
    print("\n--- Reverse Proxy Deployment Script ---\n")
    # Step 1: Ask which applications to deploy.
    selected_apps = get_applications()
    print("\nYou selected:")
    for num in selected_apps:
        print(f"  {num}) {APP_OPTIONS[num]}")

    # Step 2: Ask for backend IP and base domain.
    backend_ip = input("\nEnter the backend IP: ").strip()
    base_domain = input("Enter the base domain (e.g., example.com): ").strip()

    # Step 3: Determine if we need stream configuration.
    # If Wazuh (option 2) or Mailcow (option 3) is selected, include stream blocks.
    include_stream = any(app in selected_apps for app in [2, 3])
    if include_stream:
        print("\nStream configuration will be included (for Wazuh/Mailcow).")
    else:
        print("\nStream configuration is not required.")

    # Step 4: Check for required TLS certificate files.
    check_certificates(base_domain, selected_apps)

    # Step 5: Write out the .env file so docker-compose can use these values.
    write_env_file(backend_ip, base_domain, include_stream)

    # (Optional) You could modify or select specific docker-compose files based on the selections.
    # For simplicity, we assume your docker-compose.yml is set up to include stream blocks
    # when INCLUDE_STREAM is set to "yes".

    # Step 6: Deploy the docker-compose stack.
    deploy_stack()

    # Optionally, you can offer to tail logs.
    tail = input("Would you like to tail the logs? (y/N): ").strip().lower()
    if tail == "y":
        try:
            subprocess.run("docker-compose logs -f", shell=True)
        except KeyboardInterrupt:
            print("\nLog tailing interrupted.")
    else:
        print("Deployment complete.")


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as cpe:
        print(f"Command failed: {cpe}", file=sys.stderr)
        sys.exit(1)
