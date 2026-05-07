1. High level design - done
2. User service & Kong integration - done
3. Room service
    3.1. HTTP APIs - done
    3.2. Kong integration - pending
    3.3. Subscribe Saga - pending
4. Booking service
    4.1. Model design - done
    4.2. HTTP APIs - done
    4.3. Kong integration - pending
    4.4. Publish/subscribe Saga - pending
5. Payment service
6. Notify service

7. Local deploy
8. VPS/AWS deploy
9. CICD
10. Upgrade
    - Monitoring
    - CDC + outbox pattern
    - Idempotence
    - Performance test

11. Apply AI to gen FE
12. Re-deploy

13. Refactor project folders structure



Next: 4.4 --> 3.3
    - Setup Kafka docker
    - Booking service - Kafka config both Publish/subscribe
    - Kong - Kafka plugin to send msg to Booking
    - Room service - Kafka config to subscribe
