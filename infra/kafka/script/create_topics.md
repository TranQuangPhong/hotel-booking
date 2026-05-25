# Create Kafka Topics

Run these commands inside the broker container to create the required topics.

```bash
docker exec -it broker bash
```

Then run:

```bash
# booking_created_events - Published by booking-service when a new booking is created
kafka-topics --create --topic booking_created_events --bootstrap-server broker:29092 --partitions 3 --replication-factor 1

# room_reservation_events - Published by room-service with reservation result (success/failure)
kafka-topics --create --topic room_reservation_events --bootstrap-server broker:29092 --partitions 3 --replication-factor 1

# payment_events - Published by payment-service with payment result (success/failure)
kafka-topics --create --topic payment_events --bootstrap-server broker:29092 --partitions 3 --replication-factor 1
```

Or as a one-liner from the host:

```bash
docker exec broker kafka-topics --create --topic booking_created_events --bootstrap-server broker:29092 --partitions 3 --replication-factor 1
docker exec broker kafka-topics --create --topic room_reservation_events --bootstrap-server broker:29092 --partitions 3 --replication-factor 1
docker exec broker kafka-topics --create --topic payment_events --bootstrap-server broker:29092 --partitions 3 --replication-factor 1
```

## Verify topics

```bash
docker exec broker kafka-topics --list --bootstrap-server broker:29092
```
