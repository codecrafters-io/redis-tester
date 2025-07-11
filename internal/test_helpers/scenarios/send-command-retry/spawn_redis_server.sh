#!/usr/bin/env -S python3 -u

import socket

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(('', 6379))
    sock.listen(1)

    conn, addr = sock.accept()
    with conn:
        # first read: send null bulk string
        conn.recv(1024)
        conn.send(b"$-1\r\n")

        # second read: send pong
        conn.recv(1024)
        conn.send(b"+PONG\r\n")
