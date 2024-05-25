import sys, os
from flask import Flask, Response, send_from_directory

app = Flask(__name__)

@app.route('/')
def index():
    return send_from_directory('static', 'index.html')

@app.route('/<path:path>')
def files(path):
    return send_from_directory('static', path)

@app.route('/file')
def file():
    with open(file_path, 'rb') as f:
        return Response(f.read(), mimetype='text/plain')

if __name__ == '__main__':
    file_path = sys.argv[1]
    app.run(host='127.0.0.1', port=4334)
