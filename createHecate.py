#!/usr/bin/env python3
import os
import subprocess
import sys

# Define the utilities and their corresponding script files
UTILITIES = {
    "1": ("Create Backup", "utilities/createBackup.py"),
    "2": ("Create Config Variables", "utilities/creteConfigVariables.py"),
    "3": ("Create EOS Apps", "utilities/createEosApps.py"),
    "4": ("Create http.conf", "utilities/createHttpConf.py"),
    "5": ("Create Docker Compose", "utilities/createDockerCompose.py"),
    "6": ("Create Certificates", "utilities/createCerts.py"),
    "7": ("Restore Config", "utilities/createRestore.py"),
    "q": ("Quit", None)
}

def print_menu():
    print("\n--- Deploy Hecate Utility Wrapper ---\n")
    for key, (desc, _) in UTILITIES.items():
        print(f"{key}. {desc}")
    print()

def run_script(script_path):
    # Check if the script exists
    if not os.path.exists(script_path):
        print(f"Error: {script_path} not found.")
        return

    # Execute the script with the current python interpreter.
    # Alternatively, if these scripts are executable, you can use subprocess.call([script_path])
    try:
        subprocess.run([sys.executable, script_path], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error: Script {script_path} exited with error code {e.returncode}")
    except Exception as e:
        print(f"An error occurred while running {script_path}: {e}")

def main():
    while True:
        print_menu()
        choice = input("Enter the number of the utility to run (or 'q' to quit): ").strip()

        if choice.lower() == 'q':
            print("Exiting deployHecate.py. Goodbye!")
            break

        if choice in UTILITIES:
            desc, script_path = UTILITIES[choice]
            print(f"\nRunning '{desc}' from {script_path}...\n")
            run_script(script_path)
        else:
            print("Invalid selection. Please try again.")

if __name__ == '__main__':
    main()
