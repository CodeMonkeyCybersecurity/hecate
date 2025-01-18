# Reverse proxy to a locally hosted backend cloud web app
We have set this up on docker using docker compose, but now we are transitioning to kubernetes using k3s.

This README.md will record the steps we took to transition from a docker compose environment to a k3s environment.

The transition will probably be painful but we are documenting it here because we can't be the first people doing the docker compose -> kubernetes transition and we have struggled to find good documentation for this. Hopefully someone will find this real-life case study useful.

## Do no harm 
Before we do anything , we need to backup all our current configs
```
cd $HOME
sudo cp -r hecate/ "$(date)-hecate.bak" # need sudo here because the .pem files should have limited access.
```

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
├── 1-dev/          # Contains all resources related to the Hecate development gateway.
│   ├── certs/          # Store your SSL certificate files here. 
│   │   ├── fullchain.pem
│   │   ├── privkey.pem
│   │   ├── wazuh.fullchain.pem
│   │   └── wazuh.privkey.pem
│   ├── manifests/          # Organize your Kubernetes manifests into subdirectories for better clarity and management.
│   │   ├── secrets/            # Store Kubernetes Secret manifests. 
│   │   │   ├── hecate-tls.yaml
│   │   │   └── wazuh-tls.yaml
│   │   ├── deployments/            # Store Deployment manifests for your services.
│   │   │   ├── hecate-deployment.yaml
│   │   │   └── wazuh-deployment.yaml
│   │   ├── services/           # Store Service manifests to expose your Deployments within the cluster.
│   │   │   ├── hecate-service.yaml
│   │   │   └── wazuh-service.yaml
│   │   ├── ingress/            # Store Ingress resources for HTTP/S routing managed by Traefik.
│   │   │   ├── hecate-ingress.yaml
│   │   │   └── wazuh-ingress.yaml
│   │   ├── ingressroute_tcp/           # Store IngressRouteTCP resources for TCP (Stream) routing managed by Traefik.
│   │   │   ├── wazuh-1515-ingressroute_tcp.yaml
│   │   │   └── wazuh-1514-ingressroute_tcp.yaml
│   │   └── traefik-config.yaml         # traefik-config.yaml: Optional configuration for Traefik if you need to customize its settings beyond Helm defaults.
│   ├── nginx.conf.template
│   └── Dockerfile (if needed)
├── service2/
│   └── manifests/
│       └── ... (similar structure)
├── service3/
│   └── manifests/
│       └── ... (similar structure)
├── .env.example            # Example environmental variables file. Put sensitive values in here
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

### Create secrets
Encode your .pem files in Base64:
```
cd $HOME/hecate/1-dev/certs

# Now is a good a time as ever to ensure the certs directory has the appropriate permissions assigned to it
sudo chmod 700 $HOME/hecate/1-dev/certs

# ensure correct owner, this assumes you are the main admin of these services and have the appropatie permisssions, otherwise setting them as root may be easier
sudo chown $USER: privkey.pem wazuh.privkey.pem
sudo chown $USER: fullchain.pem wazuh.fullchain.pem

# ensure appropriate permissions
chmod 600 privkey.pem wazuh.privkey.pem # these are private keys (600 - read and write for the owner only)
chmod 644 fullchain.pem wazuh.fullchain.pem # 644 (readable by everyone, writable only by the owner).

# Webpage certificates
base64 -w 0 fullchain.pem > fullchain.b64
base64 -w 0 privkey.pem > privkey.b64

# Wazuh certificates
base64 -w 0 wazuh.fullchain.pem > wazuh.fullchain.b64
base64 -w 0 wazuh.privkey.pem > wazuh.privkey.b64
```

Verify those commands worked as expected
```
ls
```
* expected output is
```
fullchain.b64  mail.fullchain.pem  privkey.b64  wazuh.fullchain.b64  wazuh.privkey.b64
fullchain.pem  mail.privkey.pem    privkey.pem  wazuh.fullchain.pem  wazuh.privkey.pem
```

Output the .b64 values (do these one at a time)
You will need these values for the next step
```
cat fullchain.b64 # this will output <base64-encoded-fullchain.pem>
cat privkey.b64 # will output <base64-encoded-privkey.pem>
cat wazuh.fullchain.b64 # <base64-encoded-wazuh.fullchain.pem>
cat wazuh.privkey.b64 # <base64-encoded-wazuh.privkey.pem>
```

### Placing TLS secrets
Based on our directory structure, TLS secret YAML files should reside in the `$HOME/hecate/1-dev/manifests/secrets/`.
```
cd $HOME/hecate/1-dev/manifests/secrets/
```

Now create hecate-tls.yaml:
```
nano hecate-tls.yaml > EOF "
# hecate-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: hecate-tls
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-fullchain.pem>
  tls.key: <base64-encoded-privkey.pem>
EOF
```
* make sure to replace the value placeholders such as <base64-encoded-fullchain.pem> with their corresponding values output in the previous step 

Now create wazuh-tls.yaml:
```
nano hecate-tls.yaml > EOF "
# wazuh-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: wazuh-tls
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-wazuh.fullchain.pem>
  tls.key: <base64-encoded-wazuh.privkey.pem>
EOF
```
* make sure to replace the values of <base64-encoded-wazuh.privkey.pem>  and <base64-encoded-wazuh.fullchain.pem> here too



