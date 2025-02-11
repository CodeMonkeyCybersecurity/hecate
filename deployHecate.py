#!/usr/bin/env python3
"""
deploy.py

A simple interactive script to deploy the reverse proxy docker-compose stack.
It asks the user which applications to deploy (e.g. static site, Wazuh, Mailcow, etc.),
the backend IP, and the base domain. All applications except for the static HTTP website
will be deployed on a subdomain with these defaults:

    Wazuh       -> delphi
    Mailcow     -> mailcow
    Umami       -> analytics
    Mattermost  -> collaborate
    Nextcloud   -> cloud
    ERPNext     -> erp
    Jellyfin    -> jellyfin
    Grafana     -> observe
    Minio       -> s3 (and also s3api)
    Jenkins     -> jenkins
    Wiki        -> wiki

If Wazuh (option 2) or Mailcow (option 3) is selected, the final configuration will include
both the stream and HTTP blocks. The script also checks that the required TLS certificate files
are present for each FQDN.

Note: Instead of using wildcard certificates, the following file names are expected:

Static site:
    fullchain.pem
    privkey.pem

For subdomains (e.g., for Wazuh):
    delphi.fullchain.pem
    delphi.privkey.pem

For Minio:
    s3.fullchain.pem, s3.privkey.pem, s3api.fullchain.pem, s3api.privkey.pem

(and similarly for the other applications)
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
    11: "Jenkins",
    12: "Wiki"
}

# Mapping from application number to the default subdomain (for FQDN)
SUBDOMAIN_MAP = {
    2: "delphi",       # Wazuh
    3: "mailcow",      # Mailcow
    4: "analytics",    # Umami
    5: "collaborate",  # Mattermost
    6: "cloud",        # Nextcloud
    7: "erp",          # ERPNext
    8: "jellyfin",     # Jellyfin
    9: "observe",      # Grafana
    10: "s3",          # Minio (primary; also check s3api)
    11: "jenkins",     # Jenkins
    12: "wiki"         # Wiki
}

# Mapping from application number to the environment variable name for its FQDN.
ENV_VAR_MAP = {
    2: "DELPHI_DOMAIN",
    3: "MAIL_DOMAIN",
    4: "ANALYTICS_DOMAIN",
    5: "COLLABORATE_DOMAIN",
    6: "CLOUD_DOMAIN",
    7: "ERP_DOMAIN",
    8: "JELLYFIN_DOMAIN",
    9: "OBSERVE_DOMAIN",
    10: "MINIO_DOMAIN",    # For Minio primary (s3)
    11: "JENKINS_DOMAIN",
    12: "WIKI_DOMAIN"
}

# Mapping from application number to the expected conf file name(s) in conf.d/servers.
# For static site (option 1), we expect "base.conf"
CONF_FILE_MAP = {
    1: "base.conf",
    2: "delphi.conf",
    3: "mailcow.conf",
    4: "analytics.conf",
    5: "collaborate.conf",
    6: "cloud.conf",
    7: "erp.conf",
    8: "jellyfin.conf",
    9: "observe.conf",
    10: ["s3.conf", "s3api.conf"],
    11: "jenkins.conf",
    12: "wiki.conf"
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

    For a static website (option 1), we expect:
      ./certs/fullchain.pem
      ./certs/privkey.pem

    For each selected application (except option 1), we expect certificate files named:
      ./certs/<subdomain>.fullchain.pem
      ./certs/<subdomain>.privkey.pem

    For Minio (option 10), we expect both for "s3" and "s3api".
    """
    missing_files = []

    # Check static site certificates if option 1 is selected.
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

def substitute_config_files(backend_ip, base_domain, include_stream):
    """
    Recursively process configuration files (e.g. in conf.d and snippets) and substitute
    any occurrences of ${BACKEND_IP}, $BACKEND_IP, ${BASE_DOMAIN}, and ${INCLUDE_STREAM} with the
    user-provided values.
    
    This function modifies files in place.
    """
    dirs_to_process = ["conf.d", "snippets"]
    substitutions = {
        "${BACKEND_IP}": backend_ip,
        "$BACKEND_IP": backend_ip,
        "${BASE_DOMAIN}": base_domain,
        "$BASE_DOMAIN": base_domain,
        "${INCLUDE_STREAM}": "yes" if include_stream else "no",
        "$INCLUDE_STREAM": "yes" if include_stream else "no"
    }
    
    for d in dirs_to_process:
        for root, dirs, files in os.walk(d):
            for file in files:
                # Process only files ending in .conf (or you can adjust the pattern)
                if file.endswith(".conf"):
                    path = os.path.join(root, file)
                    try:
                        with open(path, "r") as f:
                            content = f.read()
                        for key, val in substitutions.items():
                            content = content.replace(key, val)
                        with open(path, "w") as f:
                            f.write(content)
                        print(f"Substituted variables in {path}")
                    except Exception as e:
                        print(f"Error processing {path}: {e}")

def cleanup_conf_files(selected_apps, include_stream):
    """
    Delete .conf files in conf.d/servers and conf.d/stream that are not used
    based on the selected applications.
    """
    # Cleanup HTTP server configuration files.
    servers_dir = os.path.join(".", "conf.d", "servers")
    expected_files = set()
    # For static site (option 1), expect "base.conf"
    if 1 in selected_apps:
        expected_files.add(CONF_FILE_MAP[1])
    # For every other selected app, add the expected file(s)
    for app in selected_apps:
        if app == 1:
            continue
        conf_entry = CONF_FILE_MAP.get(app)
        if conf_entry:
            if isinstance(conf_entry, list):
                for fname in conf_entry:
                    expected_files.add(fname)
            else:
                expected_files.add(conf_entry)
    print("Expected server conf files:", expected_files)
    for filename in os.listdir(servers_dir):
        if filename.endswith(".conf"):
            if filename not in expected_files:
                full_path = os.path.join(servers_dir, filename)
                print(f"Deleting unused server conf file: {full_path}")
                try:
                    os.remove(full_path)
                except Exception as e:
                    print(f"Error deleting {full_path}: {e}")
    
    # Cleanup stream configuration files.
    stream_dir = os.path.join(".", "conf.d", "stream")
    if include_stream:
        # Only keep stream files for Wazuh (option 2) and Mailcow (option 3) if selected.
        expected_stream_files = set()
        if 2 in selected_apps:
            expected_stream_files.add("delphi.conf")
        if 3 in selected_apps:
            expected_stream_files.add("mailcow.conf")
        print("Expected stream conf files:", expected_stream_files)
        for filename in os.listdir(stream_dir):
            if filename.endswith(".conf"):
                if filename not in expected_stream_files:
                    full_path = os.path.join(stream_dir, filename)
                    print(f"Deleting unused stream conf file: {full_path}")
                    try:
                        os.remove(full_path)
                    except Exception as e:
                        print(f"Error deleting {full_path}: {e}")
    else:
        # Remove all stream conf files if stream configuration is not needed.
        for filename in os.listdir(stream_dir):
            if filename.endswith(".conf"):
                full_path = os.path.join(stream_dir, filename)
                print(f"Deleting stream conf file (not needed): {full_path}")
                try:
                    os.remove(full_path)
                except Exception as e:
                    print(f"Error deleting {full_path}: {e}")

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
    
    # Step 5: Substitute environment variables in the configuration files.
    substitute_config_files(backend_ip, base_domain, include_stream)
    
    # Step 6: Delete unused .conf files from conf.d/servers and conf.d/stream.
    cleanup_conf_files(selected_apps, include_stream)
    
    # Step 7: Deploy the docker-compose stack.
    deploy_stack()
    
    # Optionally, offer to tail logs.
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
