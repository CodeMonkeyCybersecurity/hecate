#!/bin/bash

# Script: generate_nginx_conf.sh
# Description: Prompts the user for backendIP, HOSTNAME, and BASE_DOMAIN,
#              then replaces the placeholders in nginx.conf with these values.

# Exit immediately if a command exits with a non-zero status
set -e

# Function to prompt for input and ensure it's not empty
prompt_input() {
    local var_name=$1
    local prompt_message=$2
    local input

    while true; do
        read -rp "$prompt_message: " input
        if [[ -n "$input" ]]; then
            echo "$input"
            return
        else
            echo "Error: $var_name cannot be empty. Please enter a valid value."
        fi
    done
}

echo "=== NGINX Configuration Generator ==="

# Prompt the user for input values
backendIP=$(prompt_input "backendIP" "Enter the backend IP address")
HOSTNAME=$(prompt_input "HOSTNAME" "Enter the hostname for the NGINX server")
BASE_DOMAIN=$(prompt_input "BASE_DOMAIN" "Enter the base domain for your services")

# Define file paths
TEMPLATE_FILE="nginx.conf.template"  # Assuming you have a template file
OUTPUT_FILE="nginx.conf"

# Check if the template file exists
if [[ ! -f "$TEMPLATE_FILE" ]]; then
    echo "Error: Template file '$TEMPLATE_FILE' not found in the current directory."
    exit 1
fi

# Backup existing nginx.conf if it exists
if [[ -f "$OUTPUT_FILE" ]]; then
    cp "$OUTPUT_FILE" "${OUTPUT_FILE}.bak"
    echo "Backup of existing '$OUTPUT_FILE' created as '${OUTPUT_FILE}.bak'."
fi

# Replace placeholders with actual values using sed
sed -e "s/\${backendIP}/$backendIP/g" \
    -e "s/\${HOSTNAME}/$HOSTNAME/g" \
    -e "s/\${BASE_DOMAIN}/$BASE_DOMAIN/g" \
    "$TEMPLATE_FILE" > "$OUTPUT_FILE"

echo "nginx.conf has been generated successfully with the provided values."
