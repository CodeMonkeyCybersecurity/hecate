#!/bin/sh
set -e

# List of variables to substitute
env_vars="backendIP HOSTNAME BASE_DOMAIN"

# Substitute environment variables in the template
envsubst "${env_vars}" < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

# Optionally, you can validate the config before starting NGINX
nginx -t

# Start NGINX in the foreground
exec nginx -g 'daemon off;'
