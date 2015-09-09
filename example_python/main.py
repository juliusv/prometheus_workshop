#!/usr/bin/python

from server import Server
from client import start_request_workers


if __name__ == '__main__':
    Server().start()
    start_request_workers()
