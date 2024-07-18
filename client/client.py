import base64
import json
import logging
import requests
from datetime import datetime

logging.basicConfig(level=logging.INFO)

class Config:
    def __init__(self, consumer_key, consumer_secret, initiator_name, security_credential, shortcode, environment):
        self.consumer_key = consumer_key
        self.consumer_secret = consumer_secret
        self.initiator_name = initiator_name
        self.security_credential = security_credential
        self.shortcode = shortcode
        self.environment = environment

class MPesa:
    def __init__(self, config):
        self.config = config

    def get_api_base_url(self):
        if self.config.environment == "production":
            return "https://api.safaricom.co.ke"
        return "https://sandbox.safaricom.co.ke"

    def get_access_token(self):
        url = f"{self.get_api_base_url()}/oauth/v1/generate?grant_type=client_credentials"
        logging.info(f"Requesting access token from URL: {url}")
        
        response = requests.get(
            url,
            headers={
                "Authorization": "Basic " + base64.b64encode(f"{self.config.consumer_key}:{self.config.consumer_secret}".encode()).decode()
            }
        )

        logging.info(f"Response status code: {response.status_code}")
        logging.info(f"Response headers: {response.headers}")
        logging.info(f"Response text: {response.text}")

        if response.status_code != 200:
            logging.error(f"Failed to get access token: {response.text}")
            raise Exception("Failed to get access token")

        access_token = response.json().get("access_token")
        if not access_token:
            raise Exception("No access token in response")
        
        logging.info(f"Got access token: {access_token[:10]}")  # Log first 10 characters of token for security
        return access_token

    def initiate_b2c_payment(self, phone_number, amount, command_id="SalaryPayment", remarks="Test remarks", occasion=""):
        url = f"{self.get_api_base_url()}/mpesa/b2c/v1/paymentrequest"
        originator_conversation_id = "f02a760b-d233-45b6-9dc3-f41b49028299"

        request_body = {
            "InitiatorName": self.config.initiator_name,
            "SecurityCredential": self.config.security_credential,
            "CommandID": command_id,
            "Amount": str(amount),
            "PartyA": self.config.shortcode,
            "PartyB": phone_number,
            "Remarks": remarks,
            "QueueTimeOutURL": "https://5c19-102-216-154-4.ngrok-free.app/b2c/queue",
            "ResultURL": "https://5c19-102-216-154-4.ngrok-free.app/b2c/result",
            "Occasion": occasion,
            "OriginatorConversationID": originator_conversation_id
        }

        logging.info(f"M-Pesa B2C request body: {json.dumps(request_body)}")
        
        token = self.get_access_token()
        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }
        
        response = requests.post(url, headers=headers, data=json.dumps(request_body))
        if response.status_code != 200:
            logging.error(f"Failed to initiate M-Pesa B2C transaction: {response.text}")
            raise Exception("Failed to initiate M-Pesa B2C transaction")
        
        response_data = response.json()
        logging.info(f"M-Pesa B2C API response: {json.dumps(response_data)}")

        if response_data.get("ResponseCode") != "0":
            raise Exception(f"M-Pesa B2C transaction failed: {response_data.get('ResponseDescription')}")
        
        return response_data.get("ConversationID")

if __name__ == "__main__":
    config = Config(
        consumer_key="fG0ParfkWPYRqhCPARPN40pDzxC4sQxXfGSCYYKnuRODhEXo",
        consumer_secret="SVlF1T7luJDHpkw8Kz9Q671lpPAgWjcuC76pLdTgPdWKSg3AwYuDEvRPhEH2meNP",
        initiator_name="testapi",
        security_credential="RpTFNNdRgAFGjYd+Gdq0wHZYdvVgDbdYBNMetNw83tTSwOZTYEV8pmaAZxqtMxAFKCjoNJWNUIxj5o4T9lx6svj9FxjcV8M9J1E8q8dIqdSSeQYNBMPw+v3DTnPh9tYTzOQaVo2Rd6rctXwWhp9p8WSvZ3LvizGD9/8xcPJxouMTCb2O9GGmRT4aMkDryJKSm6D6pkVu/mDVOHc5cwfcjruU3kiYiuIcDKlFkMirObV6KQ5us/KU0QbqDRWuaX3oEGf/sHViuQJJImTu4hvypAe2b5IRsdbTqhF7RU6VzoO1QDOlCMRY90ZPkX7Pd57HqYhF+llinJw0AHWn6XSgjQ==",
        shortcode="600997",
        environment="sandbox"  # or "production"
    )

    service = MPesa(config)

    phone_number = "254708374149"
    amount = 10

    try:
        conversation_id = service.initiate_b2c_payment(phone_number, amount)
        print(f"B2C Payment initiated successfully. ConversationID: {conversation_id}")
    except Exception as e:
        logging.error(f"Error initiating B2C Payment: {e}")
