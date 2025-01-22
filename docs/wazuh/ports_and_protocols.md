# Ports and protocols
Understanding why are certain Wazuh ports are managed within the `stream` blocks in your NGINX configuration, and why are some managed in the `http` block, is crucial for setting up a secure and efficient reverse proxy. Let’s break down the reasoning behind handling port `55000` in the `http` block versus `1514` and `1515` in the stream block, and explore ways to confirm this configuration without relying solely on trial and error.

## Understanding the Nature of Each Port and Its Protocol

### Ports 1514 and 1515: Raw TCP Traffic
Usage:
 * `1514/TCP`: Primarily used for Wazuh agent communication with the Wazuh manager.
 * `1515/TCP`: Used for automatic agent enrollment requests.
 * Protocol:
 * These ports handle raw TCP traffic using Wazuh’s proprietary communication protocols. They are not based on HTTP/S protocols.
 * NGINX Configuration:
 * Since the traffic is raw TCP, these ports are appropriately managed within the stream block of NGINX, which is designed for layer 4 (TCP/UDP) proxying.

### Port 55000: HTTP/S Traffic for Wazuh API
Usage:
 * `55000/TCP`: Dedicated to the Wazuh server API, facilitating communication for tasks like agent enrollment via the API.

Protocol:
 * The Wazuh API operates over HTTP/S, adhering to RESTful principles. This involves layer 7 (application layer) protocols, which are inherently different from raw TCP traffic.
 * NGINX Configuration:
 * Given that port `55000` utilizes HTTP/S, it should be managed within the http block of NGINX, which is optimized for layer 7 proxying. This allows for functionalities like SSL termination, header manipulation, and more.

## Why Port 55000 Belongs in the http Block
API-Based Communication:
 * The Wazuh API leverages HTTP methods (`GET`, `POST`, `PUT`, `DELETE`) for communication, making it a natural fit for the http block. This allows NGINX to handle HTTP-specific features like SSL/TLS encryption, URL routing, and authentication mechanisms.
 * Enhanced Security and Functionality:
 * Managing the API within the http block enables you to implement security measures such as rate limiting, access controls, and detailed logging, which are more straightforward with HTTP/S traffic.
 * Consistency with Web Services:
 * APIs generally follow web service paradigms, aligning with the capabilities provided by the http block for efficient handling and scalability.

## Confirming the Correct Configuration Without Trial and Error
To ensure that your NGINX configuration aligns with the protocols used by each port, you can follow these verification steps:

### Refer to Official Wazuh Documentation
Wazuh Documentation:
 * The Wazuh Official Documentation is the most reliable source. It outlines the purpose and protocol requirements for each port.
 * API Specifics: The documentation clearly states that the Wazuh API operates over HTTP/S, necessitating an `http` block configuration.

### Analyze Network Traffic
Use tcpdump or Wireshark:
 * Capture and inspect the traffic on the respective ports to identify the protocols in use.
 * Example Command:
```
sudo tcpdump -i any port 55000 -A
```
 * If the traffic shows HTTP methods (e.g., `GET`, `POST`), it confirms that the port uses HTTP/S.

### Check Service Configurations
Wazuh Manager Configuration:
 * Examine the Wazuh manager’s configuration files to see how each port is defined and what protocols they are set to use.
 * Configuration Files:
 * Typically located in `/var/ossec/etc/` or similar directories.

### Utilize NGINX’s Debug Logging
Make use of detailed Logging:
 * The `1-dev` and `2-stage` instances of the NGINX configurations include debug logging for the `http` block. Use the logs to monitor how traffic is being handled.
```
sudo tail -f /var/log/nginx/error.log /var/log/nginx/access.log
```

 * Look for patterns or errors that indicate protocol mismatches.

### Leverage Command-Line Tools
Using curl for HTTP/S Verification:
 * Attempt to make an HTTP request to port `55000`.
```
read -p "What is the domain you want to test? (eg. domain.com or cybermonkey.net.au) " TEST_DOMAIN
curl -v https://wazuh-api.${TEST_DOMAIN}:55000/
```
 * Successful HTTP responses (e.g., HTTP 200) indicate proper HTTP/S handling.

 * Using nc (Netcat) for Raw TCP Verification:
 * Test ports `1514` and `1515` for raw TCP responses.
```
read -p "What is the domain you want to test? (eg. domain.com or cybermonkey.net.au) " TEST_DOMAIN
nc -vz ${TEST_DOMAIN} 1514
nc -vz ${TEST_DOMAIN} 1515
```
 * A successful connection without HTTP protocol responses suggests raw TCP handling.

### Examine NGINX Configuration Syntax
Check for Protocol-Specific Directives:
 * Within the `http` block, directives like `proxy_set_header` are specific to HTTP/S. Their presence further confirms the appropriate handling of HTTP-based traffic.

### Consult Community Forums and Support Channels
Wazuh Community:
 * Engage with the Wazuh Community or forums to verify best practices and receive insights from other users who have implemented similar configurations.

## Practical Example and Confirmation
Given the above understanding, here’s how you can confirm the proper handling of port `55000`:

### Review Your NGINX Configuration

Ensure that port `55000` is configured within the `http` block, not the stream block. Here’s an example snippet:
```
http {
    include       mime.types;
    default_type  application/octet-stream;

    # Existing configurations...

    # Wazuh API Proxy
    server {
        listen 55000 ssl;
        server_name wazuh-api.domain.com;  # Replace with your desired subdomain

        ssl_certificate /etc/nginx/certs/wazuh-api.fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/wazuh-api.privkey.pem;

        location / {
            proxy_pass https://${backendIP}:55000;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # Optional: Redirect HTTP traffic on port 55000 to HTTPS
    server {
        listen 55000;
        server_name wazuh-api.domain.com;

        return 301 https://$host$request_uri;
    }

    # Existing configurations...
}
```
### Test API Connectivity
 * Using curl:
```
read -p "What is the domain you want to test? (eg. domain.com or cybermonkey.net.au) " TEST_DOMAIN
curl -u <username>:<password> https://wazuh-api.${TEST_DOMAIN}:55000/
```
 * A successful response indicates proper HTTP/S handling.

 * Using openssl:
```
read -p "What is the domain you want to test? (eg. domain.com or cybermonkey.net.au) " TEST_DOMAIN
openssl s_client -connect wazuh-api.${TEST_DOMAIN}:55000
```
 * Look for a successful SSL handshake, confirming HTTPS configuration.

### Monitor NGINX Logs
 * Access Logs:
```
# this command needs to be run inside the nginx `hecate` docker container, or the nginx instance spun up by the docker-compose.yml on the backend
sudo tail -f /var/log/nginx/access.log
```

 * Error Logs:
```
# this command needs to be run inside the nginx `hecate` docker container, or the nginx instance spun up by the docker-compose.yml on the backend
sudo tail -f /var/log/nginx/error.log
```
 * Observe the logs during API requests to ensure there are no errors and that traffic is being proxied correctly.

## Summary and Best Practices
Protocol Alignment:
 * Raw TCP Ports (1514, 1515): Managed within the stream block.
 * HTTP/S Ports (55000): Managed within the `http` block.

Configuration Verification:
 * Always refer to official documentation and use network tools to understand the protocols in use.
 * Security Considerations:
 * Implement SSL/TLS for all HTTP/S traffic to ensure secure communication.
 * Regularly monitor logs to detect and address any anomalies.
 * Documentation and Community Resources:
 * Leverage official documentation and community forums for guidance and troubleshooting.

By understanding the underlying protocols and purposes of each port, you can confidently configure your NGINX reverse proxy to handle Wazuh’s various services effectively. This approach minimizes the need for trial and error and ensures a robust and secure setup.
