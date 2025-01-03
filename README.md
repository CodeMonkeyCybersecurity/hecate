# Hecate 

A Gateway for the Modern Cyber Underworld

Welcome to Hecate, the ultimate reverse proxy setup powered by Caddy. Named after the ancient Greek goddess of crossroads, boundaries, and the arcane arts, Hecate stands as the gatekeeper between your infrastructure and the outside world. 


## How to use this:

Apply the YAML:
```
microk8s kubectl apply -f hecate-ingress-controller.yaml
```

Verify Deployment:
```
microk8s kubectl get pods -o wide -A
microk8s kubectl get svc
```

Access Your Site:
	•	Use http://<node-IP> to access your site.


## Debugging

Cert-Manager Functionality: Ensure cert-manager is installed, and the letsencrypt-prod secret is successfully created:
```
microk8s kubectl get secrets -n development
```

Ingress Class: Verify that the Ingress controller is using the public class:
```
microk8s kubectl describe ingressclass public
```

Ingress Controller: Confirm the Ingress controller is running and functional:
```
microk8s kubectl get pods -n ingress
```

MetalLB
```
microk8s kubectl logs -n metallb-system -l app=speaker
```

Check Node Labels
```
microk8s kubectl get nodes --show-labels
```

Check Speaker Logs. The MetalLB speaker component is responsible for assigning IPs. Check its logs:
```
microk8s kubectl logs -n metallb-system -l app=speaker
```

Check Controller Logs. The controller manages the overall configuration. Check its logs:
```
microk8s kubectl logs -n metallb-system -l app=controller
```

Keep monitoring the ingress logs to ensure there are no errors or misconfigurations:
```
kubectl logs -n ingress deployment/hecate-ingress-controller
```

Check the metallb-system configuration:
```
microk8s kubectl describe configmap config -n metallb-system
```


## Complaints, compliments, confusion and other communications:

Secure email: [git@cybermonkey.net.au](mailto:git@cybermonkey.net.au)  

Website: [cybermonkey.net.au](https://cybermonkey.net.au)

```
     ___         _       __  __          _
    / __|___  __| |___  |  \/  |___ _ _ | |_____ _  _
   | (__/ _ \/ _` / -_) | |\/| / _ \ ' \| / / -_) || |
    \___\___/\__,_\___| |_|  |_\___/_||_|_\_\___|\_, |
                  / __|  _| |__  ___ _ _         |__/
                 | (_| || | '_ \/ -_) '_|
                  \___\_, |_.__/\___|_|
                      |__/
```
