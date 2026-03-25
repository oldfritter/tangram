# Project Introduction

This project aims to integrate services distributed across the network into a production system.

Communication between service nodes relies on MQ (Message Queue).

# Workflow

1. User visits WEB interface

2. Page sends data query request to API node

3. API node wraps request data and sends it to other nodes via MQ

4. Other nodes listen to MQ, process data when received, and send results back to API node via MQ

5. API node returns data from MQ to the WEB page

# Project Structure

```
example           ----- Examples
lib/kafka         ----- Kafka as message middleware
lib/rabbitmq      ----- RabbitMQ as message middleware
lib/redis         ----- Redis as message middleware
```

# Supported Message Queues

- **Redis**: Pub/Sub mode
- **Kafka**: Producer/Consumer mode
- **RabbitMQ**: Queue and Exchange mode
