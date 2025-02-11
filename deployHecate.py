#!/usr/bin/env python3
"""
deploy.py

A simple CLI tool to deploy, tear down, or tail logs for the reverse proxy
docker-compose stack.
"""

import argparse
import subprocess
import sys

def run_command(cmd):
    """
    Run a shell command and print its output.
    If the command fails, exit the script.
    """
    try:
        # run the command; capture output and errors, and print stdout as it occurs
        result = subprocess.run(cmd, shell=True, check=True, text=True)
    except subprocess.CalledProcessError as err:
        print(f"Error executing '{cmd}':\n{err}", file=sys.stderr)
        sys.exit(1)

def deploy():
    """Deploy (or restart) the docker-compose stack."""
    print("Tearing down any existing containers...")
    run_command("docker-compose down")
    print("Starting up the reverse proxy stack in detached mode...")
    run_command("docker-compose up -d")
    print("Deployment successful!")

def teardown():
    """Stop and remove the containers."""
    print("Tearing down the reverse proxy stack...")
    run_command("docker-compose down")
    print("Teardown successful!")

def logs():
    """Tail the docker-compose logs."""
    print("Tailing logs (press Ctrl+C to exit)...")
    # Note: Using subprocess.run will stream the logs.
    try:
        subprocess.run("docker-compose logs -f", shell=True)
    except KeyboardInterrupt:
        print("\nLog tailing interrupted.")

def main():
    parser = argparse.ArgumentParser(
        description="Manage the reverse proxy deployment using docker-compose."
    )
    parser.add_argument(
        "command",
        choices=["deploy", "teardown", "logs"],
        help="Action to perform: 'deploy' to bring up, 'teardown' to stop, 'logs' to tail logs."
    )
    args = parser.parse_args()

    if args.command == "deploy":
        deploy()
    elif args.command == "teardown":
        teardown()
    elif args.command == "logs":
        logs()
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
