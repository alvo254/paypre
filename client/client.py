import pika
import json

connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
channel = connection.channel()

channel.queue_declare(queue='transactions')

message = {
    "sender": "254712345678",
    "recipient": "254787654321",
    "amount": 100.50
}

channel.basic_publish(exchange='',
                      routing_key='transactions',
                      body=json.dumps(message))
print(" [x] Sent transaction message")
connection.close()