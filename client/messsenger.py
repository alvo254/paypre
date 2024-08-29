import pika
import json

message = {
    "sender": "254708374149",
    "recipient": "254708374149",
    "amount": 1500.0
}

connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
channel = connection.channel()
channel.queue_declare(queue='transactions', durable=True)

channel.basic_publish(
    exchange='',
    routing_key='transactions',
    body=json.dumps(message),
    properties=pika.BasicProperties(
        delivery_mode=2,  # make message persistent
    ))

print(" [x] Sent 'Transaction Message'")
connection.close()