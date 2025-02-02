# Hecate

For more actively maintained knowledge base and documentation, see [**Athena**](https://wiki.cybermonkey.net.au/Hecate).

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

# Features

* Lightweight NGINX container based on the nginx:alpine image.
* Automatic HTTPS certificate generation using Certbot.
* Support for serving custom static files from the html directory.
* Automatic redirection from HTTP to HTTPS.
* Docker Compose for easy deployment and management.


* A domain name (domain.com) pointing to your server’s IP address.
* Certbot installed on your server for certificate generation.

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
