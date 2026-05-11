1. High level design - done
2. User service & Kong integration - done
3. Room service
    3.1. HTTP APIs - done
    3.2. Kong integration - pending
    3.3. Subscribe Saga - pending
4. Booking service
    4.1. Model design - done
    4.2. HTTP APIs - done
    4.3. Kong integration - done
    4.4. Kafka setup - done
    4.5. Publish/subscribe - done
5. Payment service
6. Notify service

7. Local deploy
8. VPS/AWS deploy
9. CICD
10. Upgrade
    - Refactor hexagon structure + Saga msg structure
    - Monitoring
    - CDC + outbox pattern
    - Idempotence
    - Performance test

11. Apply AI to gen FE
12. Re-deploy




Next: 4.4 --> 3.3
    - Room service - Kafka config to subscribe
    - Booking service - check room status
        + Full Saga: booking --> create emp booking --> produces to Room service --> reserves room --> produces to booking --> payment
        + 
