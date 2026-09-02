#!/usr/bin/env -S python3 -u
import errno
import socket
import time

print("hey, binding to 6379")
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

deadline = time.time() + 5
while True:
    try:
        sock.bind(("", 6379))
        break
    except OSError as e:
        if e.errno != errno.EADDRINUSE or time.time() >= deadline:
            raise
        time.sleep(0.1)

sock.listen(1)

# Handle one PING slowly, then keep listening so the follow-up bind stage
# (JM1) can connect instead of treating process exit as a crash.
while True:
    conn, _ = sock.accept()
    conn.send(b"+PONG")
    time.sleep(0.1)  # ensure the tester reads with an appropriate timeout
    conn.send(b"\r\n")
