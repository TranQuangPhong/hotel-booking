🧩 Các module & chức năng
1. Gateway (Open Source: Kong/Traefik/NGINX Ingress)
- Entry point cho toàn bộ request từ client.
- Verify JWT (stateless). Sau khi authen thành công, push request vào Kafka (topic tương ứng).
- Không giữ session, chỉ làm chốt chặn bảo mật.

2. User Service (Java + Postgres)
- Đăng ký, đăng nhập, quản lý profile.
- Phát hành JWT token.
- Lưu thông tin user trong Postgres.

3.Room Service (Go + Redis + Kafka + MongoDB)
- CRUD phòng (tạo, sửa, xoá, xem).
- Lưu dữ liệu phòng trong MongoDB.
- Consume từ Kafka topic & cập nhật trạng thái phòng:
    + room-query
	+ booking_accepted
	+ booking_declined
	+ payment_failed
	+ payment_success.

4. Booking Service (Go + Redis + Kafka + MongoDB)
- Kiểm tra tình trạng phòng (available, booked, pending).
- Xử lý đặt phòng.
- Redis giữ phòng tạm thời (TTL).
- Sau khi confirm, publish message sang Kafka topic booking_accepted.
- Nếu có lỗi (ví dụ: phòng đã hết) → publish booking_declined.
- Consume từ Kafka topic & cập nhật trạng thái booking
    + booking_check
	+ booking_request
	+ payment_failed → rollback (xoá phòng trong Redis)
	+ payment_success.

5. Payment Service (Go + Kafka + MongoDB)
- Consume từ Kafka topic booking_accepted.
- Xử lý thanh toán (mock hoặc tích hợp Stripe/PayPal).
- Publish kết quả sang Kafka topic payment_success, payment_failed.

6. Notification Service (Go + Kafka + MongoDB)
- Consume từ Kafka topic payment_success, payment_failed, booking_declined.
- Gửi email/notification cho user.
- Nếu Booking/Payment thất bại → gửi thông báo lỗi.

7. Database Layer
- Postgres: dữ liệu quan hệ (User).
- MongoDB: dữ liệu linh hoạt, document-based (Room, Booking, Payment, Notification).
- Redis: giữ phòng tạm thời, xử lý timeout, cache dữ liệu tra cứu.

🎯 Tóm lại
Gateway: authen + push vào Kafka.
Kafka: backbone cho toàn bộ giao tiếp giữa service.
User Service: phát hành JWT.
Room Service: quản lý phòng.
Booking Service: đặt phòng, Redis giữ phòng tạm, hiển thị trạng thái booking.
Payment Service: xử lý thanh toán.
Notification Service: gửi thông báo.
MongoDB: lưu dữ liệu chính.
Redis: cache phòng tạm thời, cache dữ liệu tra cứu.
