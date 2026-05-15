# room-service

1. Reservation strategy: The "Pre-generated" Calendar (Explicit) and Bucket Pattern seem to be good to me.

So I summary:
1. Pure saga choreography, let room service handle room status.
2. Use date slot-based model to store room status (include date, price, booking id).
3. "Pre-generated" Calendar to optimize logic check/update status
4. MongoDB bucket pattern to optimize data organization.

{
  "room_id": "101",
  "month": "2026-05",
  "slots": {
    "20": {"status": "BOOKED", "price": 150},
    "21": {"status": "AVAILABLE", "price": 120}
  }
}

2. Cache strategy
MongoDB Buckets + Redis Hashes (for more clear, Bitmaps only when need high performance) + Luna + PENDING (hold room) + write-through + TTL + warming-up lazy (eager if vacation season) --> goo strategies.