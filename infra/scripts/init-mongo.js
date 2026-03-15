// Kết nối tới MongoDB bằng Compass hoặc mongo shell
// Ví dụ: use event_db

// Event Service
db.events.insertOne({
  _id: "evt_demo",
  name: "Demo Concert",
  location: "Hà Nội",
  date: ISODate("2026-03-01T19:00:00Z"),
  tickets_total: 1000,
  tickets_available: 950,
  metadata: {
    category: "music",
    organizer: "Demo Organizer"
  }
});

// Booking Service
db.bookings.insertOne({
  _id: ObjectId(),
  user_id: 1,
  event_id: "evt_demo",
  tickets: 2,
  status: "CONFIRMED",
  created_at: new Date()
});

// Notification Service
db.notifications.insertOne({
  _id: ObjectId(),
  user_id: 1,
  type: "EMAIL",
  message: "Your booking for Demo Concert is confirmed!",
  status: "SENT",
  created_at: new Date()
});
