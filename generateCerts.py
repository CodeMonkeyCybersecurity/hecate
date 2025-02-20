#!/usr/bin/env python3
"""
generateCerts.py

This script performs the following:
  - Checks current Docker processes.
  - Stops the Hecate service.
  - Loads previously entered values (base domain and email) from .last_certs.conf.
  - Prompts the user for:
      • Base domain (e.g. domain.com)
      • Subdomain to configure (e.g. sub) – can be left blank after confirmation.
      • Email address (e.g. example@domain.com)
  - Combines subdomain and base domain to form the full certificate domain.
  - Runs certbot in standalone mode to obtain the certificate.
  - Verifies that the certificate exists in /etc/letsencrypt/live/<full_domain>/.
  - Prompts the user to confirm the certificate “name” (for naming the copied files).
  - Copies the certificate files into the local certs/ directory with the confirmed name.
  - Sets appropriate permissions and lists the certs/ directory.
"""

import subprocess
import os
import sys

LAST_VALUES_FILE = ".last_certs.conf"

def run_command(command, shell=False):
    """Run a command and raise an exception if it fails."""
    print(f"Running command: {' '.join(command) if not shell else command}")
    subprocess.run(command, shell=shell, check=True)

def load_last_values():
    """Load saved values (base domain and email) from LAST_VALUES_FILE, if it exists."""
    values = {}
    if os.path.isfile(LAST_VALUES_FILE):
        with open(LAST_VALUES_FILE, "r") as f:
            for line in f:
                line = line.strip()
                if not line or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                values[key.strip()] = value.strip().strip('"')
    return values

def save_last_values(values):
    """Save the provided values dictionary to LAST_VALUES_FILE."""
    with open(LAST_VALUES_FILE, "w") as f:
        for key, value in values.items():
            f.write(f'{key}="{value}"\n')

def prompt_input(prompt_message, default_val=None):
    """Prompt for input with an optional default value."""
    if default_val:
        inp = input(f"{prompt_message} [{default_val}]: ").strip()
        return inp if inp else default_val
    else:
        while True:
            inp = input(f"{prompt_message}: ").strip()
            if inp:
                return inp
            print("Input cannot be empty. Please try again.")

def prompt_subdomain():
    """
    Prompts for subdomain. Allows a blank subdomain if the user confirms.
    Returns the subdomain (which may be an empty string).
    """
    while True:
        subdomain = input("Enter the subdomain to configure (e.g. sub). Leave blank if none: ").strip()
        if subdomain == "":
            confirm = input("You entered a blank subdomain. Do you wish to continue with no subdomain? (yes/no): ").strip().lower()
            if confirm in ("yes", "y"):
                return ""
            else:
                continue
        else:
            return subdomain

def main():
    try:
        # 1. Check Docker processes and stop Hecate
        print("Checking Docker processes...")
        run_command(["docker", "ps"])
        print("Stopping Hecate...")
        run_command(["docker", "compose", "down"])

        # 2. Load previous values if available
        prev_values = load_last_values()
        base_domain = prompt_input("Enter the base domain (e.g. domain.com)", prev_values.get("BASE_DOMAIN"))
        # Use prompt_subdomain() to allow a blank subdomain with confirmation.
        subdomain = prompt_subdomain()
        mail_cert = prompt_input("Enter the contact email (e.g. example@domain.com)", prev_values.get("EMAIL"))

        # Save the entered values for future runs
        new_values = {
            "BASE_DOMAIN": base_domain,
            "EMAIL": mail_cert
        }
        save_last_values(new_values)

        # 3. Combine to form the full domain for certificate.
        # If subdomain is blank, use the base domain.
        if subdomain:
            full_domain = f"{subdomain}.{base_domain}"
        else:
            full_domain = base_domain
        print(f"\nThe full domain for certificate generation will be: {full_domain}")

        # 4. Run certbot to obtain certificate
        certbot_command = [
            "sudo", "certbot", "certonly", "--standalone",
            "-d", full_domain,
            "--email", mail_cert,
            "--agree-tos"
        ]
        run_command(certbot_command)

        # 5. Verify certificates are present
        cert_path = f"/etc/letsencrypt/live/{full_domain}/"
        print(f"Verifying that the certificates are in '{cert_path}'...")
        run_command(["sudo", "ls", "-l", cert_path])

        # 6. Change directory to $HOME/hecate and ensure certs/ exists
        hecate_dir = os.path.join(os.environ["HOME"], "hecate")
        os.chdir(hecate_dir)
        os.makedirs("certs", exist_ok=True)

        # 7. Ask user to confirm certificate name.
        # If subdomain is blank, default to base_domain.
        default_cert_name = subdomain if subdomain else base_domain
        while True:
            confirm = input(f"Use certificate name '{default_cert_name}'? (yes/no): ").strip().lower()
            if confirm in ("yes", "y"):
                cert_name = default_cert_name
                break
            elif confirm in ("no", "n"):
                cert_name = prompt_input("Enter the desired certificate name (for file naming)")
                break
            else:
                print("Please answer yes or no.")

        # 8. Copy certificate files
        source_fullchain = f"/etc/letsencrypt/live/{full_domain}/fullchain.pem"
        source_privkey = f"/etc/letsencrypt/live/{full_domain}/privkey.pem"
        dest_fullchain = f"certs/{cert_name}.fullchain.pem"
        dest_privkey = f"certs/{cert_name}.privkey.pem"

        print("Copying certificate files...")
        run_command(["sudo", "cp", source_fullchain, dest_fullchain])
        run_command(["sudo", "cp", source_privkey, dest_privkey])

        # 9. Set appropriate permissions
        print("Setting appropriate permissions on the certificate files...")
        run_command(["sudo", "chmod", "644", dest_fullchain])
        run_command(["sudo", "chmod", "600", dest_privkey])

        # 10. List the certs directory
        print("Listing the certs/ directory:")
        run_command(["ls", "-lah", "certs/"])

        # Final message
        print(f"\nYou should now have the appropriate certificates for https://{full_domain}")
        print("Next, run ./generateNginxConf.sh before restarting Hecate")
        print("finis")

    except subprocess.CalledProcessError as e:
        print(f"\nAn error occurred while executing: {e.cmd}")
        print("Exiting.")
        sys.exit(1)

if __name__ == "__main__":
    main()
