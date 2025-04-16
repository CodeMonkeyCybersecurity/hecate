# 🔐 certs/

This directory is used to store TLS/SSL certificates and private keys for services managed by Hecate.

## What goes here?

Place your certificate files here, typically:

- `*.fullchain.pem` — Your full certificate chain
- `*.privkey.pem` — The corresponding private key

## Example

For a domain like `domain.com`, you might include:

## Format

- PEM-encoded files
- Secure permissions recommended (`chmod 600` for keys)

## Notes

- These certs are usually mounted into containers or referenced in config files.
- Do **not** commit private keys to version control unless this repo is local-only and access-restricted.

---

🛡️ Keep your keys safe. If you're unsure how to generate certs, consider using Let's Encrypt via Certbot or cme.sh.


