booking-service/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point (wires everything together)
├── internal/
│   ├── delivery/                    # Inputs (Primary Adapters)
│   │   ├── http/                    # Replaces your 'handler' folder
│   │   │   ├── booking_handler.go
│   │   │   └── routes.go
│   │   └── kafka/                   # Your new Kafka consumers
│   │       ├── booking_consumer.go  # Calls service.CreateBooking()
│   │       └── event_producer.go    
│   ├── service/                     # Core Business Logic
│   │   ├── booking_service.go       # Orchestrates logic, calls repository
│   │   └── interfaces.go            # Defines Repository and Producer interfaces
│   ├── repository/                  # Outputs (Secondary Adapters)
│   │   ├── mongodb/                 
│   │   │   └── booking_repo.go      # Implements the Service's DB interface
│   │   └── redis/                   # If you add caching later
│   └── models/                      # Domain Structs & Enums (No business logic)
│       ├── booking.go               
│       └── events.go                # Kafka payload definitions
├── config/
│   └── config.go                    # Loads .env or YAML configs
├── pkg/                             # (Optional) Reusable utilities
│   └── logger/
├── go.mod
└── go.sum