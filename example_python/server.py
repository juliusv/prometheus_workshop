#!/usr/bin/python

import random
import threading
import time
from BaseHTTPServer import BaseHTTPRequestHandler
from BaseHTTPServer import HTTPServer
from SocketServer import ThreadingMixIn

start = time.time()

def generate_request_handler(average_latency_seconds, error_ratio, outage_duration_seconds):
    def f(self):
        in_outage = (time.time() - start) % (10 * outage_duration_seconds) < outage_duration_seconds
        sleep_time = max(0, random.normalvariate(average_latency_seconds, average_latency_seconds/10))
        time.sleep(sleep_time * (3 if in_outage else 1))
        if random.random() < error_ratio * (10 if in_outage else 1):
            self.send_response(500)
        else:
            self.send_response(200)
        self.end_headers()
    return f

def handler_404(self):
  self.send_response(404)

      
ROUTES = {
    ('GET', "/"): lambda self: self.wfile.write("Hello World!"),
    ('GET', "/favicon.ico"): lambda self: self.send_response(404),
    ('GET', "/api/foo"): generate_request_handler(.01, .005, 23.0),
    ('POST', "/api/foo"): generate_request_handler(.02, .02, 60.0),
    ('GET', "/api/bar"): generate_request_handler(.015, .00025, 13.0),
    ('POST', "/api/bar"): generate_request_handler(.05, .01, 47.0),
}

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
      ROUTES.get(('GET', self.path), handler_404)(self)

    def do_POST(self):
      ROUTES.get(('POST', self.path), handler_404)(self)
        
class MultiThreadedHTTPServer(ThreadingMixIn, HTTPServer):
      pass

class Server(threading.Thread):
    def run(self):
        httpd = MultiThreadedHTTPServer(('', 8081), Handler)
        httpd.serve_forever()
