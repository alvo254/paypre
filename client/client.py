from flask import Flask, request, jsonify
import pika
import psycopg2
import json

app = Flask(__name__)

# RabbitMQ setup
rabbitmq_url = 'amqp://guest:guest@localhost:5672/'
try:
    connection = pika.BlockingConnection(pika.URLParameters(rabbitmq_url))
    channel = connection.channel()
    channel.queue_declare(queue='transactions', durable=True)
except Exception as e:
    print(f"Failed to connect to RabbitMQ: {e}")
    connection = None

# Database setup
try:
    conn = psycopg2.connect("dbname=rabbitmq user=alvo password=alvo254 host=localhost port=5432")
    cursor = conn.cursor()
except Exception as e:
    print(f"Failed to connect to PostgreSQL: {e}")
    conn = None

@app.route('/b2c/queue', methods=['POST'])
def queue_timeout():
    data = request.get_json()
    print("Queue Timeout:", data)
    
    # Update RabbitMQ
    if connection:
        try:
            channel.basic_publish(
                exchange='',
                routing_key='transactions',
                body=json.dumps(data),
                properties=pika.BasicProperties(
                    delivery_mode=2,  # make message persistent
                ))
        except Exception as e:
            print(f"Failed to publish message to RabbitMQ: {e}")
    
    return jsonify({"status": "success"}), 200

@app.route('/b2c/result', methods=['POST'])
def result():
    data = request.get_json()
    print("Result:", data)
    
    sender = data.get('Sender')
    recipient = data.get('Recipient')
    amount = data.get('Amount')
    checkout_request_id = data.get('CheckoutRequestID')
    response_description = data.get('ResponseDescription')
    
    if not all([sender, recipient, amount, checkout_request_id, response_description]):
        return jsonify({"status": "failure", "reason": "Invalid data"}), 400

    # Update PostgreSQL
    if conn:
        try:
            cursor.execute("""
                INSERT INTO transactions (sender, recipient, amount, created_at, checkout_request_id, status) 
                VALUES (%s, %s, %s, NOW(), %s, %s)
            """, (sender, recipient, amount, checkout_request_id, response_description))
            conn.commit()
        except Exception as e:
            print(f"Failed to insert record into PostgreSQL: {e}")
            return jsonify({"status": "failure", "reason": "Database error"}), 500
    
    # Update RabbitMQ
    if connection:
        try:
            channel.basic_publish(
                exchange='',
                routing_key='transactions',
                body=json.dumps(data),
                properties=pika.BasicProperties(
                    delivery_mode=2,  # make message persistent
                ))
        except Exception as e:
            print(f"Failed to publish message to RabbitMQ: {e}")
    
    return jsonify({"status": "success"}), 200

if __name__ == '__main__':
    app.run(port=5000)
