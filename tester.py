#!/usr/bin/python3
"""Test server"""
import sys
import time
import socket
import struct
import random
import logging
import argparse
import msgpack
from itertools import repeat
from concurrent.futures import ThreadPoolExecutor


def get_args():
    """
    get_args just parse arguments
    """
    parser = argparse.ArgumentParser(description='Check server`s heath')
    parser.add_argument("-a", "--address", dest="address", type=str,
                        help="Server`s addres.", metavar="<address>",
                        required=True, action="store")
    parser.add_argument("-p", "--port", dest="port", type=int,
                        help="Server`s port", metavar="<port>",
                        required=True, action="store")
    parser.add_argument("-t", "--timeout", dest="timeout",
                        help="Socket timeout",
                        default=5, type=int, action="store",
                        metavar='<timeout>', choices=range(60))
    parser.add_argument("-n", "--notice-number", dest="notice_number",
                        help="Notice number", metavar='<notice_number>',
                        default=10, action="store", type=int)
    parser.add_argument("-w", "--workers-number", dest="workers_number",
                        help="Workers number", metavar='<workers_number>',
                        default=5, action="store", type=int)
    parser.add_argument("--debug", dest="debug", action='store_true',
                        help="Use debug mode")
    return parser.parse_args()


def init_logger(name):
    """
    init_logger just init logger
    """
    log_format = "%(asctime)s [%(levelname)s] <%(name)s> %(message)s"
    log_level = logging.DEBUG if CONFIG.debug else logging.INFO
    logging.basicConfig(level=log_level, format=log_format)
    return logging.getLogger(str(name))


def random_domain():
    """Generate a random domain name"""
    domain = ""
    letters = "abcdefghijklmnopqrstuvwxyz"
    figures = "0123456789"
    levels = random.randint(2, 3)
    for level in range(levels):
        if level != levels - 1:
            domain += ''.join(random.choice(letters+figures)
                              for i in range(random.randint(2, 9))) + '.'
        else:
            domain += ''.join(random.choice(letters)
                              for i in range(random.randint(2, 3)))
    return domain


def random_ip():
    """Generate a random ip in string and int"""
    ip_arr = [random.randint(0, 255) for i in range(4)]
    ip_string = ".".join([str(i) for i in ip_arr])
    ip_int = struct.unpack("!I", socket.inet_aton(ip_string))[0]
    return ip_int, ip_string


class notice(object):
    __slots__ = [
        "domain",
        "ip",
        "ip_string",
    ]

    def __init__(self):
        self.domain = random_domain()
        self.ip, self.ip_string = random_ip()


    def marshal(self):
        data = {"ip": self.ip, "domain": self.domain}
        return msgpack.packb(data)


    def __str__(self):
        return "{} - {} ({})".format(self.domain, self.ip_string, self.ip)


    def __repr__(self):
        return "{} - {}".format(self.domain, self.ip_string)


def send(address, port, timeout, send_number):
    """Send notice to server"""
    logger = init_logger(send_number)
    logger.debug("Worker started")
    try:
        with socket.create_connection((address, port), timeout=timeout) as sender:
            logger.debug("Socket created :: %s", sender.dup)
            notif = notice()
            logger.info("Notice :: %s", notif)
            data = notif.marshal()
            logger.debug("Notice marshaled (%d) :: %s", len(data), str(data))
            sended = sender.send(data)
            logger.debug("Notice sended :: %d/%d", sended, len(data))
    except Exception as err:
        logger.error("Exeption :: %s", str(err))


def main():
    """Main logic of tester"""
    logger = init_logger("main")
    logger.info("The script was run with parameters: %s",
                " ".join(sys.argv))
    logger.info("Arguments parse persistently. Run workers.")
    with ThreadPoolExecutor(max_workers=CONFIG.workers_number) as executor:
        executor.map(send, repeat(CONFIG.address), repeat(CONFIG.port),
                     repeat(CONFIG.timeout), range(1, CONFIG.notice_number+1))
    logger.info("Success")


if __name__ == "__main__":
    CONFIG = get_args()
    main()
