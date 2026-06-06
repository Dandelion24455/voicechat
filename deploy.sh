#!/bin/bash
# Run this on your Mac:
#   chmod +x deploy.sh && ./deploy.sh

set -e

echo "=== Building tarball ==="
cd "$(dirname "$0")"
tar czf /tmp/voicechat-deploy.tar.gz server/ client/ mediasoup-server/ docker-compose.prod.yml .gitignore

echo "=== Uploading to server ==="
scp /tmp/voicechat-deploy.tar.gz ubuntu@212.64.28.112:/tmp/

echo "=== Rebuilding on server ==="
ssh -t ubuntu@212.64.28.112 "cd /opt/voicechat && sudo tar xzf /tmp/voicechat-deploy.tar.gz && sudo docker compose -f docker-compose.prod.yml up -d --build"

echo "=== Done! Visit http://212.64.28.112:8080 ==="
