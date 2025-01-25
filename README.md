# Hecate
A Gateway for the Modern Cyber Underworld

Welcome to Hecate, the ultimate reverse proxy setup, powered by Docker and Nginx. Named after the ancient Greek goddess of crossroads, boundaries, and the arcane arts, Hecate stands as the gatekeeper between your infrastructure and the outside world.


# Reverse proxy framework for locally hosted, cloud native backend web apps

## How do I use this repository?
This repository is best used alongside the our [`eos`](https://github.com/CodeMonkeyCybersecurity/eos.git) repository 

This project sets up an [NGINX](https://nginx.org/en/) web server as a reverse proxy using Docker Compose. The aim is to make deploying cloud native Web Apps on your own infrastructure as 'point and click' as possible. The reverse proxies set up here can be used in front of the corresponding backend web application deployed in the [`eos`](https://github.com/CodeMonkeyCybersecurity/eos.git) repository.

```
                         ┌───────────────────────────┐
                         │         Clients           │    # This is how your cloud instance will be
                         │ (User Browsers, Apps, etc)│    # accessed. Usually a browser on a client
                         └────────────┬──────────────┘    # machine.
                                      │
                                      ▼
                         ┌───────────────────────────┐
                         │       DNS Resolution      │    # This needs to be set up
                         │ (domain.com,              |    # with your cloud provider or DNS broker, eg.
                         | cybermonkey.net.au, etc.) │    # GoDaddy, Cloudflare, Hetzner, etc.
                         └────────────┬──────────────┘
                                      │
                                      ▼
            **This your remote server (reverse proxy/proxy/cloud instance)**
                          #########################
                          #  Hecate sets this up  #
                          #########################
                           ┌─────────────────┐    
                           │   Reverse Proxy │    # This is what we are setting up `hecate`.
                           │ (NGINX, Caddy,  │    # All your traffic between the internet and
                           │   Ingress, etc) │    # the backend servers gets router through
                           └───┬─────────────┘    # here.
                               │
      ┌────────────────────────┼─────────────────────────┐
      │                        │                         │
      ▼                        ▼                         ▼
      **These are your local servers (backend/virtual hosts)**
                    #######################
                    #  Eos sets these up  #
                    #######################
┌──────────────┐       ┌──────────────┐          ┌──────────────┐
│  Backend 1   │       │  Backend 2   │          │  Backend 3   │
│  (backend1)  │       │  (backend2)  │          │  (backend3)  │    # If using tailscale,
│  ┌────────┐  │       │  ┌────────┐  │          │  ┌────────┐  │    # these are the magicDNS hostnames.
│  │ Service│  │       │  │ Service│  │          │  │ Service│  │    # For setting up a demo website instance, 
│  │ Pod/   │  │       │  │ Pod/   │  │          │  │ Pod/   │  │    # see our `helen` repository
│  │ Docker │  │       │  │ Docker │  │          │  │ Docker │  │    # To set up Wazuh, check out
│  │  (eg.  │  │       │  │  (eg.  │  │          │  │  (eg.  │  │    # eos/legacy/wazuh/README.md.
│  │Website)│  │       │  │ Wazuh) │  │          │  │Mailcow)│  │    #
│  └────────┘  │       │  └────────┘  │          │  └────────┘  │    #
└──────────────┘       └──────────────┘          └──────────────┘    #
```

## Features

* Lightweight NGINX container based on the nginx:alpine image.
* Automatic HTTPS certificate generation using Certbot.
* Support for serving custom static files from the html directory.
* Automatic redirection from HTTP to HTTPS.
* Docker Compose for easy deployment and management.


* A domain name (domain.com) pointing to your server’s IP address.
* Certbot installed on your server for certificate generation.

## Supported web applications

This section outlines what cloud-native web applications are currently supported. Those currently marked :x: means they aren't supported yet and we are getting to them as one at a time.

| Web application | Hecate             | Eos                | What is does                     |
| ----------------| ------------------ | ------------------ | -------------------------------- |
| Wazuh           | :white_check_mark: | :white_check_mark: | [XDR / SIEM](https://wazuh.com/) |
| HTML websites   | :white_check_mark: | :white_check_mark: | Basic website                    |
| Mattermost      | :x:                | :x:                | [Slack alternative](https://mattermost.com/) |
| Nextcloud       | :x:                | :x:                | [iCloud /OneDrive alternative](https://nextcloud.com/) |
| Mailcow         | :x:                | :x:                | [Email/groupware](https://mailcow.email/) |
| Jenkins         | :x:                | :x:                | [CI/CD](https://jenkins.io/) |
| Grafana         | :x:                | :x:                | [Observability/monitoring](https://grafana.com/) |
| ELK Stack       | :x:                | :x:                | [Search logs/metrics](https://www.elastic.co/) |
| OpenStack       | :x:                | :x:                | [Cloud infrastructure](https://www.openstack.org/) |
| Nebula          | :x:                | :x:                | [Distributed mesh network](https://github.com/slackhq/nebula) |
| Security Onion  | :x:                | :x:                | [Security monitoring](https://securityonionsolutions.com/) |
| Restic API      | :x:                | :x:                | [Backups](https://restic.net/) |
| Keycloak        | :x:                | :x:                | [Identity and access management](https://www.keycloak.org/) |
| Theia        | :x:                | :x:                | [IDE](https://theia-ide.org/) |
| Matomo        | :x:                | :x:                | [Privacy focussed web analytics](https://matomo.org/) |
| MinIO        | :x:                | :x:                | [S3 compatible object storage](https://min.io/) |
| Penpot        | :x:                | :x:                | [UX design](https://github.com/penpot/penpot) |

More to come regarding distributed, highly available, and kubernetes-based deployments.


## Project Structure
The directory structure is important to note.

This is what the highest level in the respoitory looks like. 
```
├── 1-dev
│   ├── ...
├── 2-stage
│   ├── ...
├── 3-prod
│   ├── ...
├── 4-sh
│   ├── ...
```

### Production cycle
Each of these directories stands for the different stages of the production cycle:
1. Development
2. Staging
3. Production
4. (Optional, but common) admin/internal/command and control

We highly recommend not deploying a web app just straight to your production environment. We also appreciate that each environment likely has different configurations such as domain/hostnames, IP addresses, servers, etc. Each stage has a different directory so, after cloning the repository, you can configure the appropriate environment variables in a more modular way.

We recommend first deploying your chosen web application in your development environment first, then staging, then production, then for admin/command and control use. This is so you can gradually debug/harden your application as appropriate. While we do our best to ensure each application deployment configuration comes with sensible defaults, the internet is a wild place and each environment is different so its best to test not in a production environment. 

We encourage you to deploy each application in each environment for at least one-to-two months before graduating it to the next environment. For example, make sure your development Nextcloud instance is up and running and properly debugged for at least one-to-two months before deploying your Nextcloud to your staging instance. 


#### Admin/internal instances
The reason we recommend deploying the admin/internal instances last is because, we believe, your internal environment should be the most secured/least likely to have implmentation bugs. You can't run a production environment without your own admin panels, CI/CD, monitoring or backup servers.

The `4-sh` instance is optional because sometimes admin panels etc don't need to be exposed to the internet. The internet is a hostile environment so don't expose anything there you don't absolutely need to. 

A good example of when to use this fourth environment could be if you offer an application as a service to clients but also want to operate that service internally. For example, if you offer Grafana access as a service for clients via the cloud (in a `3-prod` production/external deployment) but also want to monitor your own infrastructure using Grafana, the 4-sh admin/internal deployment can be used.

### Each environment
```
├── docker-compose.yml         # Docker Compose configuration file
├── html/                      # Directory for your static website files (if applicable)
│   └── index.html             # Example HTML file  (if applicable)
├── nginx.conf                 # Custom NGINX configuration file
├── .env.template              # .env template, to be filled out with your variables
├── certs/                     #  Directory for SSL certificates (auto-generated)
└── README.md                  # Documentation for the setup
```

## Setup Instructions
For these instructions, a remote cloud-based front end proxy/reverse proxy server. To set up the corresponding backend 'worker' server, see [eos](https://github.com/CodeMonkeyCybersecurity/eos)

### What do I need before I get started?
* A DNS domain name, 
* The ability to configure **sub-domains** of this (eg. mail.domain.com, wazuh.domain.com). One for **each** application deployed. Each application will be accessed by the sub-domain you assign it.
* Admin access on two Ubuntu instances:
  * One, a cheap Ubuntu cloud instance with a public IP address, and the appropriate A, AAAA, CNAME and TXT and MX (if installing Mailcow) records pointing your domain and subdomains
  * A local Ubuntu server to act as a backend 'worker' node
* A VPN or other network connecting your remote cloud instance to the computer you want running your local virtualhost backend. You can set this up faily painlessly using something like [wireguard](https://www.wireguard.com/) or [tailscale](https://tailscale.com/).
* Docker and Docker Compose installed on your server 
* Certbot installed on your reverse proxy server

### Isn't it easier to just deploy the whole app on one node?
There are several reasons why we have split the deployment of the web app into two roles:
* to keep cloud costs to a minimum by running all the heavy workloads on your own computers/servers
* to not connect your home network to the internet by making sure all traffic designed for your website/web app is proxied through your cloud instance reverse proxy. If this is done correctly, this will mean that the only part of your setup directly exposed to the internet is the part controlled by the cloud provider.
* if you end up having to scale or change your infrastructure, having a reverse proxy already set up means transitioning it to being a load balancer, high availability, etc. will be much easier.

See the diagram above for clarification on how this separation of infrastructure works 


### 1. Clone the Repository

Clone this repository to your server:
```
git clone codeMonkeyCybersecurity/hecate
cd $HOME/hecate/1-dev
```

Below is a simple, reliable approach to obtain SSL certificates with Certbot and use them in an NGINX Docker container—without battling volume-mount issues for Let’s Encrypt directories. This method involves two separate steps:
1.	Use Certbot on the host (outside of Docker) to obtain certificates.

2.	Mount the certificates into your Dockerized NGINX.

By doing it this way, you avoid dealing with /var/lib/letsencrypt or /etc/letsencrypt inside Docker. Once you have your certificates on the host, you simply share them with the NGINX container.

## 1.	Stop Any Services on Port 80
Stop or remove any containers or services (like NGINX) that are currently listening on port 80:
```
docker-compose down
sudo systemctl stop nginx
```
This is necessary because Certbot’s standalone mode needs to bind port 80.

## 2.	Install Certbot on the Remote Host
On Ubuntu/Debian:
```
sudo apt update
sudo apt install certbot
```

## 3.	Obtain the Certificates (Standalone Mode)
Run Certbot to generate certificates using its built-in standalone server:
```
sudo certbot certonly --standalone \
    -d domain.com \
    --email <you>@<your.email> \
    --agree-tos
```
This will spin up a temporary web server on port 80. Certbot will place certificates in /etc/letsencrypt/live/domain.com/.

### Certificates for sub-domains
Each application you intend to install will be served on its own subdomain. Each subdomain will need its own TLS certificate. Below are examples for requesting certificates for Wazuh and Mailcow, but these can be generalised to include `any.domain.com` for any Web app you chose.

### Example 1: If you're adding Wazuh 
Run Certbot to generate certificates using its built-in standalone server:
```
sudo certbot certonly --standalone \
    -d wazuh.domain.com \
    --email <you>@<your.email> \
    --agree-tos
```

### Example 2: If you're adding Mailcow 
Run Certbot to generate certificates using its built-in standalone server:
```
sudo certbot certonly --standalone \
    -d mail.domain.com \
    --email <you>@<your.email> \
    --agree-tos
```

## 4.	Verify Certificate Files
After a successful run, check:
```
sudo ls -l /etc/letsencrypt/live/domain.com/
```

If you are deploying sub-domains, do this for each of them too. For example:
```
sudo ls -l /etc/letsencrypt/live/wazuh.domain.com/
sudo ls -l /etc/letsencrypt/live/mail.domain.com/
...
sudo ls -l /etc/letsencrypt/live/any.domain.com/
```

In each directory, you should see:
* cert.pem
* chain.pem
* fullchain.pem
* privkey.pem


## 5.	Create a Local Directory for Docker
Make a local directory in your project for the certs:
```
cd $HOME/hecate/1-dev
mkdir -p certs
```
Copy your certificates into it:
```
cd $HOME/hecate/1-dev
sudo cp /etc/letsencrypt/live/domain.com/fullchain.pem certs/
sudo cp /etc/letsencrypt/live/domain.com/privkey.pem certs/
```

Adjust permissions to be readable:
```
cd $HOME/hecate/1-dev
sudo chmod 644 certs/fullchain.pem
sudo chmod 600 certs/privkey.pem
```

### Example 1: If you're adding Wazuh 
Copy your certificates into it:
```
cd $HOME/hecate/1-dev
sudo cp /etc/letsencrypt/live/wazuh.domain.com/fullchain.pem certs/wazuh.fullchain.pem
sudo cp /etc/letsencrypt/live/wazuh.domain.com/privkey.pem certs/wazuh.privkey.pem
```

Adjust permissions to be readable:
```
cd $HOME/hecate/1-dev
sudo chmod 644 certs/wazuh.fullchain.pem
sudo chmod 600 certs/wazuh.privkey.pem
```

### Example 2: If you're adding Mailcow 
Copy your certificates into it:
```
cd $HOME/hecate/1-dev
sudo cp /etc/letsencrypt/live/mail.domain.com/fullchain.pem certs/mail.fullchain.pem
sudo cp /etc/letsencrypt/live/mail.domain.com/privkey.pem certs/mail.privkey.pem
```

Adjust permissions to be readable:
```
cd $HOME/hecate/1-dev
sudo chmod 644 certs/mail.fullchain.pem
sudo chmod 600 certs/mail.privkey.pem
```


## 6.	Use the Certificates in Docker

### For a webpage
In your docker-compose.yml, mount the local certs folder into the container and point to the copied certs in /etc/nginx/certs:
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
    restart: always
```

## 7.	Configure nginx.conf
This is where the actual custom configuration NGINX will adopt is set 

### For a webpage
```
worker_processes auto;

events {
    worker_connections 1024;
}

http {
    include mime.types;
    default_type application/octet-stream;

    error_log /var/log/nginx/error.log debug;
    access_log /var/log/nginx/access.log;

    server {
        listen 80 default_server;
        server_name ${SERVER_NAMES};

        # Redirect all HTTP traffic to HTTPS
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl default_server;
        server_name ${SERVER_NAMES};

        ssl_certificate /etc/nginx/certs/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/privkey.pem;

        location / {
            proxy_pass http://${BACKEND_IP}:${BACKEND_PORT};
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```


### Use Environment Variables
To keep sensitive values like the backend IP, port number, and hostnames confidential while using Docker Compose, you can use environment variables and a template engine to dynamically inject these values into your nginx.conf. Here’s how you can achieve that securely:

Store sensitive values in an .env file and reference them in your docker-compose.yaml. By default, these will not be committed to git as `*.env` has been added to the `.gitignore`

We have created an example env.template for you to use. For each application that you deploy, you need to delete the comment out of the file. The web page configuration comes uncommented by default, but the variables for your specific environment (ie. frontend IP, backend IP, domain and hostnames, ports) will still need to be manually set by you. 

To set your specific environment variables 
```
cd $HOME/hecate/1-dev
nano .env
```

A sample from the .env file looks like 
```
# .env
BACKEND_IP=<backend IP> # must be reachable from INSIDE the hecate docker container. If using tailscale, will look something like: 100.xxx.yyy.zzz)
BACKEND_PORT=<backend port> # must be reachable from INSIDE the hecate docker container, eg. 8080)
SERVER_NAMES=localhost <proxy-hostname> <DNS name> # eg. if using tailscale, this will look something like 'localhost domain-com domain.com'
```

Once you have set the variables you want, you need to rename the the env.template file
```
mv env.template .env
```

Examples of templates for each application could include and their corresponding nginx.conf could include:

### Example 1: If you're adding Wazuh 
```
...
    #--------------------------------------------------
    # 2) WAZUH: wazuh.domain.com
    #--------------------------------------------------
    server {
        listen 80;
        server_name wazuh.domain.com;
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl;
        server_name wazuh.domain.com;

        ssl_certificate /etc/nginx/certs/wazuh.fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/wazuh.privkey.pem;

        # Proxy pass to Kibana interface on local Wazuh
        location / {
            proxy_pass https://ww.xx.yy.zz:5601/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
...
```

A completed .env example might look something like 
```
# .env
BACKEND_IP=100.xxx.yyy.zzz
BACKEND_PORT=8080
SERVER_NAMES=localhost wazuh-domain-com wazuh.domain.com
```

Make sure the `proxy_pass https://ww.xx.yy.zz:5601/;` IP address values given above are the local backend server's tailscale IP address.


### Example 2: If you're adding mailcow
```
...
worker_processes  auto;

events {
    worker_connections  1024;
}

...

###
# STREAM BLOCK for mail protocols
###
stream {
    # Upstream definitions: mail services on the local backend
    upstream mailcow_imap_ssl {
        server ww.xx.yy.zz:993;  # IMAP-SSL on the local mailcow
    }
    upstream mailcow_smtp_tls {
        server ww.xx.yy.zz:587;  # SMTP submission on the local mailcow
    }
    # If you want to handle port 25 or 465, you can define them similarly, e.g.:
    # upstream mailcow_smtp25 {
    #     server ww.xx.yy.zz:25;
    # }

    # Listen IMAP over SSL externally
    server {
        listen 993 ssl;
        proxy_pass mailcow_imap_ssl;

        # SSL cert for mail.domain.com
        ssl_certificate /etc/nginx/certs/mail.domain.com.fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/mail.domain.com.privkey.pem;

        # (Optional) SSL Settings
        ssl_protocols       TLSv1.2 TLSv1.3;
        ssl_ciphers         HIGH:!aNULL:!MD5;
    }

    # Listen SMTP submission with STARTTLS externally
    server {
        listen 587 ssl;  # Or if you prefer to do pure TLS on 465, use 465
        proxy_pass mailcow_smtp_tls;

        ssl_certificate /etc/nginx/certs/mail.domain.com.fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/mail.domain.com.privkey.pem;

        ssl_protocols       TLSv1.2 TLSv1.3;
        ssl_ciphers         HIGH:!aNULL:!MD5;
    }

    # (Optional) If you want to forward plain 25 to Mailcow, or do SSL on 465:
    # server {
    #     listen 25;
    #     proxy_pass mailcow_smtp25;
    # }

    # Wazuh streams
    upstream wazuh_manager_1515 {
        server ww.xx.yy.zz:1515;
    }
    server {
        listen 1515;
        proxy_pass wazuh_manager_1515;
    }

    upstream wazuh_manager_1514 {
        server ww.xx.yy.zz:1514;
    }
    server {
        listen 1514;
        proxy_pass wazuh_manager_1514;
    }
}

###
# HTTP BLOCK for Web UI (Mailcow Admin, Wazuh Kibana, Static Site)
###
http {
    include       mime.types;
    default_type  application/octet-stream;

    #--------------------------------------------------
    # 1) MAIN WEBSITE: domain.com
    #--------------------------------------------------
    # Redirect HTTP → HTTPS
    server {
        listen 80;
        server_name domain.com;
        return 301 https://$host$request_uri;
    }

    # The HTTPS server for domain.com
    server {
        listen 443 ssl;
        server_name domain.com;

        ssl_certificate /etc/nginx/certs/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/privkey.pem;

        location / {
            root /usr/share/nginx/html;
            index index.html;
        }
    }

...

    #--------------------------------------------------
    # 3) MAILCOW WEB UI: mail.domain.com
    #--------------------------------------------------
    # - We do HTTP → HTTPS
    server {
        listen 80;
        server_name mail.domain.com;
        return 301 https://$host$request_uri;
    }

    # - We do HTTPS termination and pass traffic to the Mailcow web container
    server {
        listen 443 ssl;
        server_name mail.domain.com;

        ssl_certificate /etc/nginx/certs/mail.domain.com.fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/mail.domain.com.privkey.pem;

        location / {
            proxy_pass http://ww.xx.yy.zz:8080; 
            # ^ Adjust if your Mailcow web UI is mapped differently,
            #   for example: "http://ww.xx.yy.zz:80" if you published it on 80 inside Docker.

            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
...
```

## 8.	Start NGINX
With certificates in place and nginx.conf updated, start your container:
```
docker-compose down
docker-compose up -d
```

You should now test your endpoints. Using a **private browsing window**, navigate to:

### For your website
* http://domain.com/ → should redirect to HTTPS.
* https://domain.com/ → should load your static page.

### Example 1: If you're adding Wazuh 
* https://wazuh.domain.com/ → should proxy to Wazuh.

### Example 2: If you're adding mailcow
* https://mail.domain.com/ → should load the Mailcow interface.

Each of the web applications listed will be accessible via the relevant subdomain so **make sure this is set up in your DNS provider**


## Securing the setup 
Below are a few important security considerations:

* Always use long, unique, complex passphrases for all user accounts
* Keep your system up to date
* Run regular backups and test restores 

### Firewall
Allow inbound on 80/443 (for web) + the mail ports (993 and 587, 25 if needed), 1514 and 1515 for Wazuh.

It is highly recommended to make sure you have a WAF up, such as ModSecurity, on top of ufw.

```
sudo ufw status
```

For all deployments
```
sudo ufw allow http 
sudo ufw allow https 
```

Example 1: for Wazuh
```
sudo ufw allow 1514
sudo ufw allow 1515
sudo ufw allow 5601
sudo ufw allow 55000
sudo ufw allow 9200
```

Example 2: for Mailcow
```
sudo ufw allow 25
sudo ufw allow 587
sudo ufw allow 993
```

### Fail2Ban
* Set up fail2ban on each of your proxy servers. They are internet facing and so are therefore will be almost certainly constantly scraped, probed, pinged or credential stuffed. Doing what you can to limit brite force attacks is a good idea
* Set up Fail2Ban on the Mailcow server to monitor Dovecot (IMAP) and Postfix (SMTP) logs. This prevents brute-force login attempts.
* You can also run Fail2Ban on the remote proxy, but typically for mail specifically, Fail2Ban is most effective on the actual mail server that has the logs.

### TLS Ciphers
* In your NGINX config, specify strong ciphers:
```
ssl_protocols       TLSv1.2 TLSv1.3;
ssl_ciphers         EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH;
ssl_prefer_server_ciphers on;
```
* Disable weak protocols, etc.

### DNS & SPF/DKIM/DMARC
Specifically for Mailcow
* Make sure you publish correct SPF records pointing to your server that will send mail.
* Enable DKIM in Mailcow’s admin interface.
* Publish a DMARC record (optional but recommended).


# Complaints, compliments, confusion:

Secure email: [git@cybermonkey.net.au](mailto:git@cybermonkey.net.au)  

Website: [cybermonkey.net.au](https://cybermonkey.net.au)

```
#
#     ___         _       __  __          _
#    / __|___  __| |___  |  \/  |___ _ _ | |_____ _  _
#   | (__/ _ \/ _` / -_) | |\/| / _ \ ' \| / / -_) || |
#    \___\___/\__,_\___| |_|  |_\___/_||_|_\_\___|\_, |
#                  / __|  _| |__  ___ _ _         |__/
#                 | (_| || | '_ \/ -_) '_|
#                  \___\_, |_.__/\___|_|
#                      |__/
#
```
