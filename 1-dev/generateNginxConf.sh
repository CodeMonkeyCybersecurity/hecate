#!/bin/bash
# generateNginxConf.sh

# Exit immediately if a command exits with a non-zero status
set -e

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    exit 1
fi

# Export environment variables from .env
export $(grep -v '^#' .env | xargs)

# Define paths
TEMPLATE=/etc/nginx/nginx.conf.template
OUTPUT=/etc/nginx/nginx.conf

# Check if the template exists
if [ ! -f "$TEMPLATE" ]; then
    echo "Error: nginx.conf.template not found at $TEMPLATE"
    exit 1
fi

# Substitute environment variables in the template
envsubst '${backendIP} ${HOSTNAME} ${BASE_DOMAIN}' < "$TEMPLATE" > "$OUTPUT"

echo "nginx.conf has been generated successfully."

# Validate the NGINX configuration
nginx -t

# Start NGINX in the foreground
exec nginx -g 'daemon off;'
