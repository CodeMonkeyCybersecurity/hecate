#!/bin/bash

# Script: generate_nginx_conf.sh
# Description: Prompts the user for backendIP, HOSTNAME, and BASE_DOMAIN,
#              then replaces placeholders in nginx.conf.template with these values.
#              Remembers last-used values in .last_nginx_conf.

set -e

# Where we store the last-used values
LAST_VALUES_FILE=".last_nginx_conf"

# If we have a saved file from a previous run, load it.
# This defines $backendIP, $HOSTNAME, $BASE_DOMAIN if present.
if [[ -f "$LAST_VALUES_FILE" ]]; then
    # shellcheck disable=SC1090
    source "$LAST_VALUES_FILE"
fi

# Function to prompt for input with optional default
prompt_input() {
    local var_name=$1
    local prompt_message=$2
    # Current default (possibly loaded from $LAST_VALUES_FILE) 
    # We use indirect expansion: ${!var_name} refers to the value of the variable whose name is in $var_name
    local default_val=${!var_name}  
    local input

    while true; do
        # Show [default] if we have one
        if [[ -n "$default_val" ]]; then
            read -rp "$prompt_message [$default_val]: " input
        else
            read -rp "$prompt_message: " input
        fi

        if [[ -z "$input" && -n "$default_val" ]]; then
            # If user pressed Enter with no new input, keep the old default
            echo "$default_val"
            return
        elif [[ -n "$input" ]]; then
            # User typed something new, use that
            echo "$input"
            return
        else
            # The user left it empty and there's no default to fall back on
            echo "Error: $var_name cannot be empty. Please enter a valid value."
        fi
    done
}

echo "=== NGINX Configuration Generator ==="

# Prompt for values (will show defaults if any)
backendIP=$(prompt_input "backendIP" "Enter the backend IP address")
HOSTNAME=$(prompt_input "HOSTNAME" "Enter the hostname for the NGINX server")
BASE_DOMAIN=$(prompt_input "BASE_DOMAIN" "Enter the base domain for your services")

# Save the values so future runs start with the same defaults
cat <<EOF > "$LAST_VALUES_FILE"
backendIP="$backendIP"
HOSTNAME="$HOSTNAME"
BASE_DOMAIN="$BASE_DOMAIN"
EOF

TEMPLATE_FILE="nginx.conf.template"
OUTPUT_FILE="nginx.conf"

# Check if template file exists
if [[ ! -f "$TEMPLATE_FILE" ]]; then
    echo "Error: Template file '$TEMPLATE_FILE' not found in the current directory."
    exit 1
fi

# Backup existing nginx.conf if it exists
if [[ -f "$OUTPUT_FILE" ]]; then
    cp "$OUTPUT_FILE" "$(date +"%Y%m%d_%H%M%S")_${OUTPUT_FILE}.bak"
    echo "Backup of existing '$OUTPUT_FILE' created."
fi

# Replace placeholders with user values
sed -e "s/\${backendIP}/$backendIP/g" \
    -e "s/\${HOSTNAME}/$HOSTNAME/g" \
    -e "s/\${BASE_DOMAIN}/$BASE_DOMAIN/g" \
    "$TEMPLATE_FILE" > "$OUTPUT_FILE"

echo "'$OUTPUT_FILE' generated successfully using the provided values."
