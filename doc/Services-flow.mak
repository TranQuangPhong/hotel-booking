🔄 Luồng đi chính (CQRS pattern & event driven / may be upgrade to event-sourcing later)

I. Write flow (asynchronous qua Kafka)
1. Client → Gateway → Kafka
- User gửi request (login, xem sự kiện, đặt phòng).
- Gateway verify JWT. Nếu token hợp lệ, Gateway push request vào Kafka (topic tương ứng).
- Payload message gồm: userId, roles, traceId, requestData.

3. Booking Service
- Consume từ Kafka topic booking-request.
- Giữ phòng tạm trong Redis.
- Sau khi confirm, publish sang Kafka topic booking_accepted (hoặc booking_declined).

4. Payment Service
- Consume từ Kafka topic booking_accepted.
- Xử lý thanh toán.
- Publish kết quả sang Kafka topic payment_success, payment_failed.

5. Notification Service
- Consume từ Kafka topic payment_success, payment_failed.
- Gửi email/notification cho user.

6.1. Room Service
- Consume từ Kafka topic & cap nhật trạng thái phòng:
    + booking_accepted (step 3)
	+ booking_declined (step 3)
	+ payment_success (step 4)
	+ payment_failed (step 4).
6.2. Booking Service
- Consume từ Kafka topic & cập nhật trạng thái booking:
	+ payment_failed (step 4) → rollback (xoá phòng trong Redis)
	+ payment_success (step 4).

*** Luồng lỗi minh họa (Saga choreography)
- User đặt phòng → Booking giữ phòng trong Redis.
- Booking publish “booking_accepted”.
- Payment xử lý, nhưng thất bại.
- Payment publish “payment_failed”.
- Booking consume “payment_failed” → xóa phòng trong Redis.
- Notification consume “payment_failed” → gửi email báo lỗi cho user.

Rollback với Saga (Choreography qua Kafka)
- Payment Service: Publish payment_success hoặc payment_failed.
- Booking Service: Nếu nhận event payment_failed → rollback phòng trong Redis.
- Room Service: Nếu nhận event booking_declined hoặc payment_failed → cập nhật trạng thái phòng (available).
- Notification Service:
	+ Nếu payment_success → gửi email xác nhận.
	+ Nếu payment_failed → gửi email báo lỗi.


II. Read flow (synchronous rest API)

1. Room Service
- Nhan truc tiep http request.
- Trả dữ liệu từ Redis/MongoDB.
2. Booking Service
- Nhan truc tiep http request.
- Trả dữ liệu từ Redis/MongoDB.