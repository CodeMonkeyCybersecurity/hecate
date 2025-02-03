#!/bin/bash
# generateCerts.sh

echo "stop hecate, for now"
docker compose down

read -p "What subdomain do you need a certificate for? (Must be in this format: sub.domain.com) " SUB_CERT
read -p "What email can you be contacted on? (Must be in this format: you@youremail.com) " MAIL_CERT
sudo certbot certonly --standalone \
    -d $SUB_CERT \
    --email $MAIL_CERT \
    --agree-tos

echo "Verify certificates are in 'etc/letsencrypt'"
sudo ls -l /etc/letsencrypt/live/$SUB_CERT/

cd $HOME/hecate
mkdir -p certs/

read -p "What is the subdomain these certificates are for? (eg. if mail, enter 'mail', if wazuh enter 'wazuh', etc. If a base domain, leave blank) " CERT_NAME

sudo cp /etc/letsencrypt/live/$SUB_CERT/fullchain.pem certs/$CERT_NAME.fullchain.pem
sudo cp /etc/letsencrypt/live/$SUB_CERT/privkey.pem certs/$CERT_NAME.privkey.pem

echo "setting appropriate permissions"
sudo chmod 644 certs/*fullchain.pem
sudo chmod 600 certs/*privkey.pem

echo "verify certs are present"
ls -lah certs/

echo "bring Hecate back up"
docker compose up -d

echo "check docker processes"
docker ps 

echo "You should now have the appropriate certificates for https://$SUB_CERT"



echo "finis"
