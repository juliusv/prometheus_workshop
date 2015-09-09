#!/usr/bin/python

import math
import urllib2
import thread
import time

OSCILLATION_PERIOD_SECONDS = 300.0


def send_request(method, path):
    data = None
    if method == 'POST':
        data = ''
    try:
        urllib2.urlopen('http://localhost:8081' + path, data)
    except urllib2.HTTPError:
        pass
    except:
        pass

start = time.time()

def oscillation_factor():
    return 2 + math.sin(math.sin(2 * math.pi * (time.time() - start) / OSCILLATION_PERIOD_SECONDS))

def request_worker(method, path, sleep):
    while True:
        send_request(method, path)
        time.sleep(sleep * oscillation_factor())

def start_request_workers():
    thread.start_new_thread(request_worker, ('GET', '/api/foo', .01))
    thread.start_new_thread(request_worker, ('POST', '/api/foo', .15))
    thread.start_new_thread(request_worker, ('GET', '/api/bar', .02))
    thread.start_new_thread(request_worker, ('POST', '/api/foo', .1))
    thread.start_new_thread(request_worker, ('GET', '/api/nonexistent', .5))
