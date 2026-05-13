🧩 Các module & chức năng
1. Gateway (Open Source: Kong/Traefik/NGINX Ingress)
- Entry point cho toàn bộ request từ client.
- Verify JWT (stateless) + Route request.
- Không giữ session, chỉ làm chốt chặn bảo mật.

2. User Service (Java + Postgres)
- Đăng ký, đăng nhập, quản lý profile.
- Phát hành JWT token.
- Lưu thông tin user trong Postgres.

3.Room Service (Go + Redis + Kafka + MongoDB)
- CRUD phòng (tạo, sửa, xoá, xem).
- Lưu dữ liệu phòng trong MongoDB.
- Consume từ Kafka topic & update trạng thái phòng:
	+ booking_created → tạm giữ phòng trong Redis (TTL), update trạng thái phòng thành pending.
	+ payment_events -> update trạng thái phòng (đã bán / hủy bán).
- Publish message: topic room_reservation_events sau khi check status / or reservation.

4. Booking Service (Go + Redis + Kafka + MongoDB)
- CRUD booking (tạo, sửa, xoá, xem).
- Tạo booking order -> publish message: topic booking_created.
- Consume từ Kafka topic & update trạng thái booking
	+ room_reservation_events -> tạo booking chính thức.
	+ payment_events -> finalize booking (SUCCESS/FAILED).

5. Payment Service (Go + Kafka + MongoDB)
- Consume từ Kafka topic room_reservation_events -> xử lý thanh toán (mock hoặc tích hợp Stripe/PayPal).
- Publish kết quả: topic payment_events.

6. Notification Service (Go + Kafka + MongoDB)
- Consume từ Kafka topic room_reservation_events -> notify nếu reservation fails (ví dụ: phòng đã hết).
- Consume từ Kafka topic payment_events -> notify kết quả thanh toán (thành công/thất bại).

7. Database Layer
- Postgres: dữ liệu quan hệ (User).
- MongoDB: dữ liệu linh hoạt, document-based (Room, Booking, Payment, Notification).
- Redis: giữ phòng tạm thời, xử lý timeout, cache dữ liệu tra cứu.

🎯 Tóm lại
Gateway: authen + route request.
Kafka: backbone cho toàn bộ giao tiếp giữa service.
User Service: phát hành JWT.
Room Service: quản lý phòng, reservation (source of truth)
Booking Service: create/update trạng thái booking.
Payment Service: xử lý thanh toán.
Notification Service: gửi thông báo.
MongoDB: lưu dữ liệu chính.
Redis: cache phòng tạm thời, cache dữ liệu tra cứu.
