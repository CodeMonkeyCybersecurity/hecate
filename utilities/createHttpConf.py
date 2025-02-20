#!/usr/bin/env python3
import os
import shutil

def prompt_with_default(prompt, default, description):
    print(f"\n{description}")
    user_input = input(f"{prompt} [{default}]: ").strip()
    return user_input if user_input else default

def main():
    config_file = "http.conf"
    backup_file = "http.conf.bak"

    print("Welcome to the HTTP block configuration updater for your http.conf file.")
    print("Below you'll see a description for each setting along with its current default value.")
    print("Press Enter to keep the default value or type a new one as desired.\n")
    
    # Prompt for configuration options with descriptions and sensible defaults
    server_tokens = prompt_with_default(
        "Hide NGINX version (server_tokens)",
        "off",
        "server_tokens: When set to 'off', NGINX will not display its version in error pages and headers."
    )
    
    include_mime = prompt_with_default(
        "Path to MIME types file (include)",
        "mime.types",
        "include: This file defines MIME types for various file extensions."
    )
    
    default_type = prompt_with_default(
        "Default MIME type (default_type)",
        "application/octet-stream",
        "default_type: The default MIME type for files with unknown extensions."
    )
    
    error_log_path = prompt_with_default(
        "Error log path",
        "/var/log/nginx/error.log",
        "error_log: The file path where NGINX will log error messages."
    )
    
    error_log_level = prompt_with_default(
        "Error log level (e.g., warn, debug)",
        "warn",
        "Error log level: Determines the minimum severity of messages to be logged (e.g., 'warn' or 'debug')."
    )
    
    access_log_path = prompt_with_default(
        "Access log path",
        "/var/log/nginx/access.log",
        "access_log: The file path where NGINX will log access details for incoming requests."
    )
    
    sendfile = prompt_with_default(
        "Sendfile (on/off)",
        "on",
        "sendfile: When 'on', NGINX uses the sendfile system call to transfer files efficiently."
    )
    
    include_servers = prompt_with_default(
        "Include server blocks directory/pattern",
        "/etc/nginx/conf.d/servers/*.conf",
        "include: Specifies the path or pattern for including server block configuration files."
    )

    # Build the configuration content
    config_content = f"""###
# HTTP BLOCK
###

http {{
    # Hide NGINX version
    server_tokens   {server_tokens};
    include         {include_mime};
    default_type    {default_type};

    error_log       {error_log_path} {error_log_level};    # change warn to debug if installing a development server
    access_log      {access_log_path};       # enable access logging 
    sendfile        {sendfile};
    
    ###
    # SERVER BLOCKS
    ###
    include          {include_servers};
}}
"""

    # Check if http.conf exists and rename it to http.conf.bak
    if os.path.exists(config_file):
        # If a backup already exists, we can remove it or handle accordingly.
        if os.path.exists(backup_file):
            os.remove(backup_file)
        shutil.move(config_file, backup_file)
        print(f"\nExisting {config_file} has been backed up to {backup_file}")

    # Write the new configuration to http.conf
    with open(config_file, "w") as f:
        f.write(config_content)
    print(f"\nNew configuration has been written to {config_file}")

if __name__ == '__main__':
    main()
