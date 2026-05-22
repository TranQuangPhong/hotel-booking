#!/bin/bash
# Manual replica set init script (alternative to mongo-room-init container).
# Run inside the container: docker exec -it mongo-room-db bash /path/to/this/script
# NOTE: The mongo-room-init container in docker-compose handles this automatically.

mongosh --port 27027 -u room_service -p room_service --authenticationDatabase admin <<EOF
rs.initiate({
    _id: "rs-room",
    members: [{_id: 0, host: "mongo-room-db:27027"}]
});
rs.status();
EOF
