from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/b2c/queue', methods=['POST'])
def queue_timeout():
    data = request.get_json()
    print("Queue Timeout:", data)
    # Handle queue timeout notification
    return jsonify({"status": "success"}), 200

@app.route('/b2c/result', methods=['POST'])
def result():
    data = request.get_json()
    print("Result:", data)
    # Handle result notification
    return jsonify({"status": "success"}), 200

if __name__ == '__main__':
    app.run(ssl_context='adhoc')