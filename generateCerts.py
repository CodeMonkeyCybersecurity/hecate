#!/usr/bin/env python3
"""
generate_certs.py

This script replicates the functionality of generateCerts.sh:
  - Checks current Docker processes.
  - Stops the Hecate service.
  - Prompts the user for a subdomain and email.
  - Runs certbot in standalone mode to obtain a certificate.
  - Verifies the certificate location.
  - Copies the certificate files to the local hecate certs folder.
  - Sets appropriate permissions on the certificate files.
  - Lists the certificates directory.
  - Provides final instructions to the user.
"""

import subprocess
import os
import sys

def run_command(command, shell=False):
    """Run a command and raise an exception if it fails."""
    print(f"Running command: {' '.join(command) if not shell else command}")
    subprocess.run(command, shell=shell, check=True)

def main():
    try:
        # 1. Check Docker processes
        print("Checking Docker processes...")
        run_command(["docker", "ps"])

        # 2. Stop Hecate
        print("Stopping Hecate...")
        run_command(["docker", "compose", "down"])

        # 3. Get user input for the certificate subdomain and email
        sub_cert = input("What subdomain do you need a certificate for? (Must be in this format: sub.domain.com) ").strip()
        mail_cert = input("What email can you be contacted on? (Must be in this format: you@youremail.com) ").strip()

        # 4. Run certbot to obtain certificate
        certbot_command = [
            "sudo", "certbot", "certonly", "--standalone",
            "-d", sub_cert,
            "--email", mail_cert,
            "--agree-tos"
        ]
        run_command(certbot_command)

        # 5. Verify certificates are in /etc/letsencrypt/live/<sub_cert>/
        print("Verifying that the certificates are in '/etc/letsencrypt/live/{}/'...".format(sub_cert))
        run_command(["sudo", "ls", "-l", f"/etc/letsencrypt/live/{sub_cert}/"])

        # 6. Change directory to $HOME/hecate and ensure certs/ directory exists
        hecate_dir = os.path.join(os.environ["HOME"], "hecate")
        os.chdir(hecate_dir)
        os.makedirs("certs", exist_ok=True)

        # 7. Get user input for the certificate name (used to name the copied files)
        cert_name = input("What is the subdomain these certificates are for? (eg. if mail, enter 'mail'; if wazuh, enter 'wazuh'; if a base domain, leave blank) ").strip()

        # 8. Copy the certificate files
        source_fullchain = f"/etc/letsencrypt/live/{sub_cert}/fullchain.pem"
        source_privkey = f"/etc/letsencrypt/live/{sub_cert}/privkey.pem"
        dest_fullchain = f"certs/{cert_name}.fullchain.pem"
        dest_privkey = f"certs/{cert_name}.privkey.pem"

        print("Copying certificate files...")
        run_command(["sudo", "cp", source_fullchain, dest_fullchain])
        run_command(["sudo", "cp", source_privkey, dest_privkey])

        # 9. Set appropriate permissions on the certificate files
        print("Setting appropriate permissions on the certificate files...")
        run_command(["sudo", "chmod", "644", dest_fullchain])
        run_command(["sudo", "chmod", "600", dest_privkey])

        # 10. Verify that the certs are present
        print("Listing the certs/ directory:")
        run_command(["ls", "-lah", "certs/"])

        # Final messages
        print(f"\nYou should now have the appropriate certificates for https://{sub_cert}")
        print("Next, run ./generateNginxConf.sh before restarting Hecate")
        print("finis")

    except subprocess.CalledProcessError as e:
        print(f"\nAn error occurred while executing: {e.cmd}")
        print("Exiting.")
        sys.exit(1)

if __name__ == "__main__":
    main()
