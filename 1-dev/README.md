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

## Directory structure
We need to reorganize project structure
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

## NGINX -> Traefik
Migrate our NGINX Configuration to Traefik because this is k3s' default Gateway
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

### Secrets: Handle TLS Certificates with Kubernetes Secrets
#### Transcribe our certs so they can be used by Traefik
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

#### Placing TLS secrets so they can be used 
Based on our directory structure, TLS secret YAML files should reside in the `$HOME/hecate/1-dev/manifests/secrets/`.
```
cd $HOME/hecate/1-dev/manifests/secrets/
```

Now create hecate-tls.yaml:
```
cat <<EOF > hecate-tls.yaml
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
cat <<EOF > wazuh-tls.yaml
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


#### Applying the Secrets to our K3S Cluster
Apply the Secrets:
```
cd $HOME/hecate/1-dev/manifests/secrets/
kubectl apply -f hecate-tls.yaml
kubectl apply -f wazuh-tls.yaml
```

Verify they have been applied correctly
```
kubectl get secrets
```

The output should look something like
```
NAME         TYPE                DATA   AGE
hecate-tls   kubernetes.io/tls   2      <age>
wazuh-tls    kubernetes.io/tls   2      <age>
```

#### Make sure not to expose our secrets to Git publically
Open `.gitignore`:
```
cd $HOME/hecate
nano .gitignore
```

Add the following lines to .gitignore file to exclude the secrets directory and .env files:
```
# Kubernetes Secrets
hecate/1-dev/manifests/secrets/*.yaml
hecate/2-stage/manifests/secrets/*.yaml
hecate/3-prod/manifests/secrets/*.yaml
hecate/4-sh/manifests/secrets/*.yaml

# Environment variables
.env
```

#### We decided just for our development environment not to use External Secret Management just now
For enhanced security, we considered using external secret management tools like HashiCorp Vault, Sealed Secrets, or External Secrets Operator because these tools provide more robust mechanisms for managing and distributing secrets securely.



### Deployments and Services
Now secrets are defined, we can define deployments and services

#### Deployments

Reverse proxy for [Helen](https://github.com/CodeMonkeyCybersecurity/helen.git). To create the reverse proxy for our plain HTML webpage, [Helen](https://github.com/CodeMonkeyCybersecurity/helen.git)

Navigate to the appropriate directory
```
cd $HOME/hecate/1-dev/manifests/deployments
```

##### Hecate Web Deployment `hecate-deployment.yaml`:
Create the hecate deployment:
```
cat << EOF > hecate-deployment.yaml
# hecate-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hecate
  labels:
    app: hecate
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hecate
  template:
    metadata:
      labels:
        app: hecate
    spec:
      containers:
      - name: hecate
        image: nginx:latest
        ports:
        - containerPort: 80
        - containerPort: 443
        - containerPort: 1515
        - containerPort: 1514
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx/nginx.conf
          subPath: nginx.conf
        - name: certs
          mountPath: /etc/nginx/certs
      volumes:
      - name: nginx-config
        configMap:
          name: hecate-nginx-config
      - name: certs
        secret:
          secretName: hecate-tls
EOF
```
 
##### Hecate Wazuh Deployment `wazuh-deployment.yaml`:
For our Wazuh instance [delphi](https://github.com/CodeMonkeyCybersecurity/eos/tree/main/legacy/wazuh). Assuming Wazuh backend is accessible on ports 5601.
```
cat << EOF > wazuh-deployment.yaml
# wazuh-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wazuh
  labels:
    app: wazuh
spec:
  replicas: 1
  selector:
    matchLabels:
      app: wazuh
  template:
    metadata:
      labels:
        app: wazuh
    spec:
      containers:
      - name: wazuh
        image: wazuh/wazuh:latest
        ports:
        - containerPort: 5601
        # Add necessary environment variables or volumes if required
EOF
```

#### Services
Navigate to the appropriate directory
```
cd $HOME/hecate/1-dev/manifests/services/
```
##### Hecate Web Service `hecate-service.yaml`:
Create hecate service. This will expose Hecate internally within the cluster.
```
cat <<EOF > hecate-service.yaml
# hecate-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: hecate-service
  labels:
    app: hecate
spec:
  selector:
    app: hecate
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
    - protocol: TCP
      port: 443
      targetPort: 443
    - protocol: TCP
      port: 1515
      targetPort: 1515
    - protocol: TCP
      port: 1514
      targetPort: 1514
  type: ClusterIP
```

##### Hecate Wazuh Service `wazuh-service.yaml`:
```
cat << EOF > wazuh-service.yaml
# wazuh-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: wazuh-service
  labels:
    app: wazuh
spec:
  selector:
    app: wazuh
  ports:
    - protocol: TCP
      port: 5601
      targetPort: 5601
  type: ClusterIP
EOF
```

### Traefik for Ingress and TCP Routing
Traefik, as an Ingress Controller, can handle both HTTP/S and TCP (Stream) traffic efficiently. Below, I’ll outline how to configure Traefik to replicate your NGINX reverse proxy setup.

#### HTTP/S Routing with Ingress Resources
Define Ingress resources to handle HTTP/S traffic. Traefik can manage SSL termination and routing based on hostnames.
Navigate to the ingress directory
```
cd $HOME/hecate/1-dev/manifests/ingress/
```

##### Hecate Web Ingress `hecate-ingress.yaml`:
```
cat << EOF > hecate-ingress.yaml
# hecate-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hecate-ingress
  namespace: default
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  rules:
  - host: <FQDN>
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: <backend-service-name>
            port:
              number: 8080
  tls:
  - hosts:
    - <FQDN>
    secretName: hecate-tls
EOF
```

##### Hecate Wazuh Ingress `wazuh-ingress.yaml`
```
cat << EOF > wazuh-ingress.yaml
# wazuh-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: wazuh-ingress
  namespace: default
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  rules:
  - host: wazuh.cybermonkey.dev
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: wazuh-service
            port:
              number: 5601
  tls:
  - hosts:
    - wazuh.cybermonkey.dev
    secretName: wazuh-tls
EOF
```

#### TCP (Stream) Routing with IngressRouteTCP
Traefik allows you to define TCP services via IngressRouteTCP resources.

Navigate to the `ingressroute_tcp` directory
```
cd $HOME/hecate/1-dev/manifests/ingressroute_tcp/
```

Create `hecate-ingressroute_tcp.yaml`
```
cat << EOF > hecate-ingressroute_tcp.yaml
# hecate-ingressroute_tcp.yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: wazuh-1515
  namespace: default
spec:
  entryPoints:
    - wazuh1515
  routes:
    - match: HostSNI(`*`)
      services:
        - name: wazuh-service
          port: 1515

---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: wazuh-1514
  namespace: default
spec:
  entryPoints:
    - wazuh1514
  routes:
    - match: HostSNI(`*`)
      services:
        - name: wazuh-service
          port: 1514
EOF
```

Define Traefik EntryPoints for TCP Ports `traefik-config.yaml`:
```
cat << EOF > traefik-config.yaml
# traefik-config.yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: tcp-entrypoints
  namespace: default
spec:
  entryPoints:
    - web
    - websecure
    - wazuh1515
    - wazuh1514
EOF
```

### Understanding our Current Traefik Setup in K3S
Since we installed the basic K3S distribution, which includes Traefik by default, we need to adjust the existing Traefik configuration differently to add additional entrypoints for TCP (Stream) traffic on ports 1515 and 1514.

We need to go through the necessary steps to modify Traefik’s configuration within our K3S cluster to support our Hecate-dev gateway requirements.

K3S comes bundled with Traefik as the default Ingress Controller, installed in the kube-system namespace. To customize Traefik (e.g., adding new entrypoints for TCP traffic), we need to modify its existing configuration.

#### Verify Traefik Deployment:
Run the following command to list Traefik-related resources. Let's do this in our `$HOME`, but create a new folder so we don't make a mess:
```
cd $HOME
mkdir -p edittingTraefik
cd edittingTraefik
kubectl get all -n kube-system | grep traefik
```

Expected output (confirming that Traefik is running correctly within our cluster.):
```
...
deployment.apps/traefik-57b79cf995-rl8kw   1/1     Running   0          55s
service/traefik                          ClusterIP 10.43.232.190   <none>        9000/TCP,80/TCP,443/TCP   55s
...
```

#### Adding Additional EntryPoints to Traefik
To handle TCP (Stream) traffic on ports 1515 and 1514, you’ll need to define new entrypoints in Traefik’s configuration.

##### Edit Traefik’s ConfigMap
Traefik’s configuration in K3S is managed via a ConfigMap named traefik in the kube-system namespace. You’ll modify this ConfigMap to include the new TCP entrypoints. Retrieve the Existing ConfigMap:
```
kubectl get configmap traefik -n kube-system -o yaml > traefik-configmap.yaml
```

Let's back this up before openning the `traefik-configmap.yaml` file in your preferred text editor:
```
sudo cp traefik-configmap.yaml "$(date)-traefik-configmap.yaml.bak" # need sudo here because the .pem files should have limited access.
```

##### Edit the ConfigMap:
```
nano traefik-configmap.yaml
```

Add New EntryPoints:
Locate the entryPoints section and add the new TCP entrypoints wazuh1515 and wazuh1514. Here’s an example of how to modify it:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik
  namespace: kube-system
data:
  traefik.yaml: |
    entryPoints:
      web:
        address: ":80"
      websecure:
        address: ":443"
      wazuh1515:
        address: ":1515/tcp"
      wazuh1514:
        address: ":1514/tcp"

    providers:
      kubernetesCRD: {}
      kubernetesIngress: {}

    api:
      insecure: true
      dashboard: true

    log:
      level: DEBUG # This level of debugging is only appropraite for development servers.
```
Notes:
* Ensure proper indentation as YAML is sensitive to spaces.
* The entryPoints section now includes wazuh1515 and wazuh1514 listening on ports 1515 and 1514 respectively with the /tcp protocol.

##### Apply the Updated ConfigMap:
```
kubectl apply -f traefik-configmap.yaml
```

##### Restart Traefik Pods to Apply Changes:
After updating the ConfigMap, Traefik needs to reload its configuration. Restarting the Traefik pods ensures the new configuration is loaded.
```
kubectl rollout restart deployment/traefik -n kube-system
```

##### Verify Traefik Pods Are Running with Updated Configuration:
Ensure that the Traefik pods are in the Running state without restarts indicating issues.
```
kubectl get pods -n kube-system | grep traefik
```

#### Confirm EntryPoints Are Active

Access the Traefik dashboard to verify that the new entrypoints are active.
##### Port-Forward to Access Dashboard:
```
kubectl port-forward service/traefik -n kube-system 9000:9000
```

##### Access Dashboard:
Open your browser and navigate to http://<hostname>:9000/dashboard/.

##### Check EntryPoints:
In the dashboard, navigate to the “Entrypoints” section to confirm that wazuh1515 and wazuh1514 are listed and active.


## Apply
Applying and Verifying Your Kubernetes Manifests

Now that Traefik is configured to handle both HTTP/S and TCP traffic, proceed to apply your remaining Kubernetes manifests.

### Navigate to Your Manifests Directory:
```
cd $HOME/hecate/1-dev/manifests/
```

### Apply TLS Secrets:

Ensure you’ve already created and applied hecate-tls.yaml and wazuh-tls.yaml as discussed earlier.
```
kubectl apply -f secrets/hecate-tls.yaml
kubectl apply -f secrets/wazuh-tls.yaml
```

### Apply Deployments and Services:
#### Hecate Web Deployment and Service:
```
kubectl apply -f deployments/hecate-deployment.yaml
kubectl apply -f services/hecate-service.yaml
```

#### Hecate Wazuh Deployment and Service:
```
kubectl apply -f deployments/wazuh-deployment.yaml
kubectl apply -f services/wazuh-service.yaml
```

### Apply Ingress Resources:
#### HTTP/S Ingress for Hecate:
```
kubectl apply -f ingress/hecate-ingress.yaml
```

#### HTTP/S Ingress for Wazuh:
```
kubectl apply -f ingress/wazuh-ingress.yaml
```

### Apply IngressRouteTCP Resources:

As done previously, apply the TCP routes.
```
kubectl apply -f hecate-stage/manifests/ingressroute_tcp/ingressroute_tcp.yaml
```

### Verify All Resources Are Applied Correctly:
#### Check Deployments:
```
kubectl get deployments
```

Expected output
```
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
hecate    1/1     1            1           5m
wazuh     1/1     1            1           5m
```