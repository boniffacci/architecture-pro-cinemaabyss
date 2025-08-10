from flask import Flask, request, jsonify
from kafka import KafkaProducer, KafkaConsumer
import threading
import json
import logging
import os

logging.basicConfig(level=logging.INFO)
app = Flask(__name__)

KAFKA_BROKER = os.getenv("KAFKA_BROKERS", "kafka:9092")
TOPIC = os.getenv("KAFKA_TOPIC", "events-topic")
PORT = int(os.getenv("PORT", 8082))

producer = KafkaProducer(
    bootstrap_servers=KAFKA_BROKER,
    value_serializer=lambda v: json.dumps(v).encode("utf-8")
)


def consume():
    consumer = KafkaConsumer(
        TOPIC,
        bootstrap_servers=KAFKA_BROKER,
        auto_offset_reset="earliest",
        group_id="events-group",
        value_deserializer=lambda m: json.loads(m.decode("utf-8"))
    )
    for msg in consumer:
        logging.info(f"Consumed: {msg.value}")

threading.Thread(target=consume, daemon=True).start()


def send_event(event_type, data):
    payload = {"type": event_type, "data": data}
    producer.send(TOPIC, payload)
    producer.flush()
    logging.info(f"Produced: {payload}")


@app.route("/api/events/health", methods=["GET"])
def health_check():
    return jsonify({"status": True}), 200


@app.route("/api/events/<event_type>", methods=["POST"])
def create_event(event_type):
    if event_type not in ["user", "payment", "movie"]:
        return jsonify({"error": "Invalid event type"}), 400
    send_event(event_type.capitalize(), request.json or {})
    return jsonify({"status": "success"}), 201


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=PORT)
