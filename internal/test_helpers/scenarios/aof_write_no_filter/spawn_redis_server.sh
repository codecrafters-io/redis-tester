#!/bin/sh
# Buggy ep6: AOF dir + manifest + empty incr; appends RESP for every command (including GET, PING,
# ECHO, CONFIG GET), so the tester fails — it expects only modifying commands in the append-only file.
exec python3 -u - "$@" <<'PY'
import argparse
import os
import socket
import sys
from typing import List, Optional


def encode_resp_array(parts: List[bytes]) -> bytes:
    out = [f"*{len(parts)}\r\n".encode()]
    for p in parts:
        out.append(f"${len(p)}\r\n".encode() + p + b"\r\n")
    return b"".join(out)


def try_consume_one_array_command(buf: bytes):
    """
    If buf begins with a complete RESP array of bulk strings, return (parts, nbytes_consumed).
    Otherwise return None (need more data or invalid).
    """
    if len(buf) < 4 or buf[0:1] != b"*":
        return None
    eol = buf.find(b"\r\n", 0)
    if eol < 0:
        return None
    try:
        n = int(buf[1:eol])
    except ValueError:
        return None
    if n < 0:
        return None
    i = eol + 2
    parts = []
    for _ in range(n):
        if i >= len(buf) or buf[i : i + 1] != b"$":
            return None
        eol = buf.find(b"\r\n", i)
        if eol < 0:
            return None
        try:
            ln = int(buf[i + 1 : eol])
        except ValueError:
            return None
        i = eol + 2
        if i + ln + 2 > len(buf):
            return None
        body = buf[i : i + ln]
        i += ln
        if buf[i : i + 2] != b"\r\n":
            return None
        i += 2
        parts.append(body)
    return parts, i


def reply(parts: List[bytes], store: dict, dir_path: Optional[str]) -> bytes:
    if not parts:
        return b"-ERR empty command\r\n"
    cmd = parts[0].upper()
    if cmd == b"SET" and len(parts) == 3:
        store[parts[1]] = parts[2]
        return b"+OK\r\n"
    if cmd == b"GET" and len(parts) == 2:
        v = store.get(parts[1])
        if v is None:
            return b"$-1\r\n"
        return b"$" + str(len(v)).encode() + b"\r\n" + v + b"\r\n"
    if cmd == b"PING":
        return b"+PONG\r\n"
    if cmd == b"ECHO" and len(parts) == 2:
        e = parts[1]
        return b"$" + str(len(e)).encode() + b"\r\n" + e + b"\r\n"
    if (
        cmd == b"CONFIG"
        and len(parts) == 3
        and parts[1].upper() == b"GET"
    ):
        opt = parts[2]
        val = (dir_path or "").encode()
        return encode_resp_array([opt, val])
    return b"-ERR unknown\r\n"


def main() -> None:
    parser = argparse.ArgumentParser(add_help=False)
    parser.add_argument("--port", default="6379")
    parser.add_argument("--dir", default=None)
    parser.add_argument("--appendonly", default=None)
    parser.add_argument("--appenddirname", default=None)
    parser.add_argument("--appendfilename", default=None)
    parser.add_argument("--appendfsync", default=None)
    args, _unknown = parser.parse_known_args()

    incr_path = None

    if (
        args.dir
        and args.appenddirname
        and args.appendfilename
        and (args.appendonly or "").lower() == "yes"
    ):
        append_dir = os.path.join(args.dir, args.appenddirname)
        os.makedirs(append_dir, exist_ok=True)
        incr_basename = f"{args.appendfilename}.1.incr.aof"
        manifest_basename = f"{args.appendfilename}.manifest"
        incr_path = os.path.join(append_dir, incr_basename)
        manifest_path = os.path.join(append_dir, manifest_basename)
        with open(manifest_path, "w", encoding="utf-8") as mf:
            mf.write(f"file {incr_basename} seq 1 type i\n")
        open(incr_path, "wb").close()
        sys.stderr.write(
            f"Created {append_dir}, empty {incr_basename}, {manifest_basename} ok (no AOF filter)\n"
        )
        sys.stderr.flush()

    port = int(args.port)
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(("0.0.0.0", port))
    sock.listen()
    sys.stderr.write(f"Listening on 0.0.0.0:{port}\n")
    sys.stderr.flush()

    store = {}

    while True:
        conn, _ = sock.accept()
        buf = b""
        try:
            while True:
                chunk = conn.recv(65536)
                if not chunk:
                    break
                buf += chunk
                while True:
                    parsed = try_consume_one_array_command(buf)
                    if parsed is None:
                        break
                    parts, nbytes = parsed
                    buf = buf[nbytes:]
                    if incr_path is not None:
                        with open(incr_path, "ab") as af:
                            af.write(encode_resp_array(parts))
                    conn.sendall(reply(parts, store, args.dir))
        except OSError:
            pass
        finally:
            conn.close()


if __name__ == "__main__":
    main()
PY
