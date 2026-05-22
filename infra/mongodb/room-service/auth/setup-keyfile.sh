#!/bin/bash
# This script runs during MongoDB's default init phase (via /docker-entrypoint-initdb.d/).
# It copies the keyfile from the read-only bind mount to a persistent container path.
# The keyfile is needed for --keyFile in the command args (inter-member auth for replica set).

mkdir -p /etc/mongodb/keyfile
cp /data/keyfile-source/replica-keyfile.key /etc/mongodb/keyfile/replica-keyfile.key
chmod 400 /etc/mongodb/keyfile/replica-keyfile.key
echo "Keyfile copied to /etc/mongodb/keyfile/ and permissions set to 400"
