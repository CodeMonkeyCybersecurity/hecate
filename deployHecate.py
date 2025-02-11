#!/usr/bin/env python3
"""
deploy.py

A simple interactive script to deploy the reverse proxy docker-compose stack.
It asks the user which applications to deploy (e.g. static site, Wazuh, Mailcow, etc.),
the backend IP, and the base domain. All applications except for the static HTTP website
will be deployed on a subdomain with these defaults:

    Wazuh       -> delphi
    Mailcow     -> mail
    Umami       -> analytics
    Mattermost  -> collaborate
    Nextcloud   -> cloud
    ERPNext     -> erp
    Jellyfin    -> media
    Grafana     -> observe
    Minio       -> s3 (and also s3api)
    Jenkins     -> jenkins

If Wazuh (option 2) or Mailcow (option 3) is selected, the final configuration will include
both the stream and HTTP blocks. The script also checks that the required TLS certificate files
are present for each FQDN.

Note: Instead of using wildcard certificates, the following file names are expected:

Static site:
    fullchain.pem
    privkey.pem

For subdomains:
    delphi.fullchain.pem / delphi.privkey.pem
    mail.fullchain.pem / mail.privkey.pem
    analytics.fullchain.pem / analytics.privkey.pem
    collaborate.fullchain.pem / collaborate.privkey.pem
    cloud.fullchain.pem / cloud.privkey.pem
    erp.fullchain.pem / erp.privkey.pem
    media.fullchain.pem / media.privkey.pem
    observe.fullchain.pem / observe.privkey.pem
    s3.fullchain.pem / s3.privkey.pem
    s3api.fullchain.pem / s3api.privkey.pem
    jenkins.fullchain.pem / jenkins.privkey.pem
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

# Mapping from application number to the default subdomain (for FQDN)
SUBDOMAIN_MAP = {
    2: "delphi",       # Wazuh
    3: "mail",         # Mailcow
    4: "analytics",    # Umami
    5: "collaborate",  # Mattermost
    6: "cloud",        # Nextcloud
    7: "erp",          # ERPNext
    8: "media",        # Jellyfin
    9: "observe",      # Grafana
    10: "s3",          # Minio (for primary domain; also check s3api separately)
    11: "jenkins"      # Jenkins
}

# Mapping from application number to the environment variable name that will hold its FQDN.
ENV_VAR_MAP = {
    2: "DELPHI_DOMAIN",
    3: "MAIL_DOMAIN",
    4: "ANALYTICS_DOMAIN",
    5: "COLLABORATE_DOMAIN",
    6: "CLOUD_DOMAIN",
    7: "ERP_DOMAIN",
    8: "MEDIA_DOMAIN",
    9: "OBSERVE_DOMAIN",
    10: "MINIO_DOMAIN",    # For Minio primary (s3)
    11: "JENKINS_DOMAIN"
}


def get_applications():
    """Prompt the user to select which applications to deploy."""
    print("Select which applications to deploy behind the reverse proxy:")
    for num, name in APP_OPTIONS.items():
        print(f"  {num}) {name}")
    choices = input("Enter comma-separated numbers (e.g., 1,3,7): ")
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

    For a static website (option 1), we expect certificates for the base domain:
      ./certs/fullchain.pem
      ./certs/privkey.pem

    For each selected application (except option 1), we expect certificates for the
    corresponding subdomain using the subdomain name alone:
      ./certs/<subdomain>.fullchain.pem
      ./certs/<subdomain>.privkey.pem

    For Minio (option 10), we also expect certificates for the "s3api" subdomain.
    """
    missing_files = []

    # Check certificates for the static site (option 1)
    if 1 in selected_apps:
        static_fullchain = "./certs/fullchain.pem"
        static_privkey = "./certs/privkey.pem"
        if not os.path.exists(static_fullchain):
            missing_files.append(static_fullchain)
        if not os.path.exists(static_privkey):
            missing_files.append(static_privkey)

    # Check certificates for each other application.
    for app in selected_apps:
        if app == 1:
            continue
        # For Minio (option 10), check two sets.
        if app == 10:
            for name in ["s3", "s3api"]:
                fullchain = f"./certs/{name}.fullchain.pem"
                privkey = f"./certs/{name}.privkey.pem"
                if not os.path.exists(fullchain):
                    missing_files.append(fullchain)
                if not os.path.exists(privkey):
                    missing_files.append(privkey)
        else:
            subdomain = SUBDOMAIN_MAP[app]
            fullchain = f"./certs/{subdomain}.fullchain.pem"
            privkey = f"./certs/{subdomain}.privkey.pem"
            if not os.path.exists(fullchain):
                missing_files.append(fullchain)
            if not os.path.exists(privkey):
                missing_files.append(privkey)

    if missing_files:
        print("\nERROR: The following TLS certificate file(s) are missing:")
        for f in missing_files:
            print(f"  - {f}")
        print("Please make sure all required certificate files are available before deploying.")
        sys.exit(1)
    else:
        print("All required certificate files are present.\n")


def write_env_file(backend_ip, base_domain, selected_apps, include_stream):
    """
    Write a .env file with the provided values.
    The docker-compose file should read variables (e.g. BACKEND_IP, BASE_DOMAIN, INCLUDE_STREAM,
    and each application's subdomain FQDN) from this file.
    """
    env_lines = [
        f"BACKEND_IP={backend_ip}",
        f"BASE_DOMAIN={base_domain}",
        f"INCLUDE_STREAM={'yes' if include_stream else 'no'}"
    ]
    
    # For the static HTTP site (option 1), its domain is the base domain.
    if 1 in selected_apps:
        env_lines.append(f"STATIC_SITE_DOMAIN={base_domain}")
    
    # For each other application, compute its FQDN (subdomain + base_domain) and add an environment variable.
    for app in selected_apps:
        if app == 1:
            continue
        if app in ENV_VAR_MAP and app in SUBDOMAIN_MAP:
            fqdn = f"{SUBDOMAIN_MAP[app]}.{base_domain}"
            env_lines.append(f"{ENV_VAR_MAP[app]}={fqdn}")
            print(f"{ENV_VAR_MAP[app]} set to {fqdn}")
        # Special handling for Minio (option 10): add a variable for s3api.
        if app == 10:
            fqdn_api = f"s3api.{base_domain}"
            env_lines.append(f"MINIO_API_DOMAIN={fqdn_api}")
            print(f"MINIO_API_DOMAIN set to {fqdn_api}")
    
    env_content = "\n".join(env_lines) + "\n"
    try:
        with open(".env", "w") as env_file:
            env_file.write(env_content)
        print("\nEnvironment file (.env) written successfully.\n")
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
    
    # Step 4: Check for required TLS certificate files for each domain/subdomain.
    check_certificates(base_domain, selected_apps)
    
    # Step 5: Write out the .env file so docker-compose can use these values.
    write_env_file(backend_ip, base_domain, selected_apps, include_stream)
    
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
