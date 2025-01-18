# Reverse proxy to a locally hosted backend cloud web app
We have set this up on docker using docker compose, but now we are transitioning to kubernetes using k3s.

This README.md will record the steps we took to transition from a docker compose environment to a k3s environment.

The transition will probably be painful but we are documenting it here because we can't be the first people doing the docker compose -> kubernetes transition and we have struggled to find good documentation for this. Hopefully someone will find this real-life case study useful.

## Stopping docker services 
This is going to result in some downtime 
```
cd $HOME/hecate/1-dev
docker compose down
```

We then [installed k3s](https://github.com/CodeMonkeyCybersecurity/eos/tree/main/legacy/k3s)


## Development reorganisation
### Reorganize project structure
Our current directory structure:
```
1-dev
├── certs
│   ├── fullchain.pem
│   ├── privkey.pem
│   ├── wazuh.fullchain.pem
│   └── wazuh.privkey.pem
├── docker-compose.yaml
├── env.template
├── nginx.conf
└── README.md
```

We want to get it to look like:
```
hecate/
├── 1-dev/
│   ├── certs/
│   │   ├── fullchain.pem
│   │   ├── privkey.pem
│   │   ├── wazuh.fullchain.pem
│   │   └── wazuh.privkey.pem
│   ├── manifests/
│   │   ├── secrets/
│   │   │   ├── hecate-tls.yaml
│   │   │   └── wazuh-tls.yaml
│   │   ├── deployments/
│   │   │   ├── hecate-deployment.yaml
│   │   │   └── wazuh-deployment.yaml
│   │   ├── services/
│   │   │   ├── hecate-service.yaml
│   │   │   └── wazuh-service.yaml
│   │   ├── ingress/
│   │   │   ├── hecate-ingress.yaml
│   │   │   └── wazuh-ingress.yaml
│   │   ├── ingressroute_tcp/
│   │   │   ├── wazuh-1515-ingressroute_tcp.yaml
│   │   │   └── wazuh-1514-ingressroute_tcp.yaml
│   │   └── traefik-config.yaml
│   ├── nginx.conf.template
│   └── Dockerfile (if needed)
├── service2/
│   └── manifests/
│       └── ... (similar structure)
├── service3/
│   └── manifests/
│       └── ... (similar structure)
├── .env.example
├── .gitignore
└── README.md
```

For now, all we will do is 
```
cd $HOME/hecate/1-dev
mkdir -p manifests/
cd manifests/
mkdir -p secrets/ deployments/ ingress/ ingressroute_tcp/
```

## Migrating our NGINX Configuration to Traefik
Given that Traefik is our Ingress Controller in K3S, we’ll translate our existing nginx.conf into Kubernetes resources. Traefik handles HTTP/S traffic through Ingress or IngressRoute resources and TCP (Stream) traffic via IngressRouteTCP.

This is our current nginx.conf
```
# nginx.conf
worker_processes  auto;

events {
    worker_connections  1024;
}

# The STREAM block
stream {
    upstream wazuh_manager_1515 {
        server ${BACKEND_IP}:1515;
    }
    server {
        listen 1515;
        proxy_pass wazuh_manager_1515;
    }

    upstream wazuh_manager_1514 {
        server ${BACKEND_IP}:1514;
    }
    server {
        listen 1514;
        proxy_pass wazuh_manager_1514;
    }
}

# The HTTP block
http {

    include       mime.types;
    default_type  application/octet-stream;

    # Enable debug logging
    error_log /var/log/nginx/error.log debug;

    # enable access logging
    access_log /var/log/nginx/access.log;

    server {
        listen 80 default_server;
        server_name localhost ${HOSTNAME} ${FQDN}; # or _ for any host

        # Redirect all HTTP traffic to HTTPS
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl default_server;
        server_name localhost ${HOSTNAME} ${FQDN}; # or _ for any host

        ssl_certificate /etc/nginx/certs/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/privkey.pem;

        location / {
            proxy_pass http://${BACKEND_IP}:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    server {
        listen 80;
        server_name wazuh.${FQDN};
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl;
        server_name wazuh.${FQDN};
        ssl_certificate /etc/nginx/certs/wazuh.fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/wazuh.privkey.pem;
        location / {
            proxy_pass https://${BACKEND_IP}:5601/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

For reference, this is our current docker-compose.yaml:
```
# docker-compose.yaml
services:
  nginx:
    image: nginx
    container_name: hecate-dev
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro # Custom NGINX configuration
      - ./certs:/etc/nginx/certs:ro  # SSL certificates
    ports:
      - "80:80"
      - "443:443"
      - "1515:1515"
      - "1514:1514"
    restart: always
```

