🔄 Luồng đi chính (CQRS pattern & event driven / may be upgrade to event-sourcing later)

I. Write flow (incoming HTTP request - sync & services communication via Kafka - async)
1. Client → Gateway → services
- User gửi request (login, xem sự kiện, đặt phòng).
- Gateway verify JWT. Nếu token hợp lệ, route to services.
- Payload message gồm: userId, roles, traceId, requestData.

2. Booking Service
- Nhận request đặt phòng, create booking order.
- Publish: topic booking_created.

3. Room Service
- Consume: topic booking_created.
- Kiểm tra tình trạng phòng:
	+ Nếu phòng còn → giữ phòng trong Redis (TTL)
- Publish: topic room_reservation_events (both success & failure).
- Consume: topic payment_events -> update final room status (đã bán / hủy bán).

4.1. Payment Service
- Consume: topic room_reservation_events.
- Xử lý thanh toán.
- Publish: topic payment_events.

4.2. Booking Service
- Consume: topic & update trạng thái booking:
	+ room_reservation_events (step 3) -> create booking chính thức.
	+ payment_events (step 4.1) -> final booking (SUCCESS/FAILED).

5. Notification Service
- Consume: topic room_reservation_events (step 3) -> notify nếu reservation fails (ví dụ: phòng đã hết).
- Consume: topic payment_events (step 4.1) -> gửi email/notification cho user.


*** Luồng lỗi minh họa (Saga choreography)
- User đặt phòng → Booking create booking order.
- Booking publish “booking_created”.
- Room consume “booking_created” → kiểm tra phòng.
	+ Phòng còn → giữ phòng trong Redis, publish room_reservation_events.
- Payment xử lý, nhưng thất bại -> publish FAILED result to topic "payment_events".
- Room consume “payment_events” → rollback trạng thái phòng -> available.
- Booking consume “payment_events” → final booking status FAILED.
- Notification consume “payment_events” → gửi email báo lỗi cho user.

Rollback với Saga (Choreography qua Kafka)
- Room service: Publish “room_reservation_events” sau khi check status / or reservation.
- Payment Service: Publish “payment_events”.
- Booking Service: Nếu payment FAILED / reservation failed (room not available) → rollback booking trong DB.
- Room Service: Nếu payment FAILED → rollback trạng thái phòng (available).
- Notification Service:
	+ Nếu payment SUCCESS → gửi email xác nhận.
	+ Nếu payment FAILED → gửi email báo lỗi.


II. Read flow (synchronous rest API)

1. Room Service
- Nhan truc tiep http request.
- Trả dữ liệu từ Redis/MongoDB.
2. Booking Service
- Nhan truc tiep http request.
- Trả dữ liệu từ Redis/MongoDB.
