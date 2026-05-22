#!/bin/bash
# This script runs during MongoDB's default init phase (via /docker-entrypoint-initdb.d/).
# It copies the keyfile from the read-only bind mount to /tmp/ where the mongodb user has write access.
# /tmp/ persists for the container's entire lifecycle (survives restarts, only cleared on container removal).
# The keyfile is needed for --keyFile in the command args (inter-member auth for replica set).
# NOTE: /tmp/ is used because the mongodb user cannot mkdir in /etc/ or /data/ (only /data/db/).

cp /data/keyfile-source/replica-keyfile.key /tmp/replica-keyfile.key
chmod 400 /tmp/replica-keyfile.key
echo "Keyfile copied to /tmp/ and permissions set to 400"
