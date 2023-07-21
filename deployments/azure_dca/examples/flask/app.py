from flask import Flask
import os
secret = os.gentenv('secret-message')
app = Flask(__name__)

@app.route('/read_secret', methods=['GET'])
def read_secret():
    return secret
