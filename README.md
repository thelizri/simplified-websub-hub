# simple-websub-hub

## 🛠️ Running the Hub

To build and run the hub with Docker:

```bash
docker-compose up --build
```

## 📬 Subscribing to the Hub

### Example HTTP POST Request (Subscription)

Subscribers must send a `POST` request to `/` with form-encoded parameters:

```http
POST / HTTP/1.1
Host: hub:8080
Accept-Encoding: gzip
Content-Length: 134
Content-Type: application/x-www-form-urlencoded
User-Agent: Go-http-client/1.1

hub.callback=http%3A%2F%2Fweb-sub-client%3A8080%2FzMHFRyvdPf&hub.mode=subscribe&hub.secret=OrOTXImZufSUiFQkZrRm&hub.topic=%2Fa%2Ftopic
```

### Parameters

- `hub.callback` – URL where the subscriber wants to receive updates
- `hub.mode` – must be `subscribe`
- `hub.secret` – HMAC secret for verifying messages
- `hub.topic` – topic the subscriber is subscribing to

---

## 🔐 Message Signing

All messages sent to subscribers are JSON and include a signature:

```
X-Hub-Signature: sha256=<HMAC-SHA256 signature>
```

The hub computes the signature using the subscriber's secret and the message body.

---

## Subscriber Client Features

You can interact with the subscriber client (modfin/websub-client) using the following endpoints:

| Endpoint     | Description                                     |
| ------------ | ----------------------------------------------- |
| `GET /log`   | View valid messages the subscriber has received |
| `GET /resub` | Retry the subscription process (the "dance")    |

### Example Commands

```bash
# Check received messages
curl http://localhost:8081/log

# Retry the subscription request
curl http://localhost:8081/resub
```

---
