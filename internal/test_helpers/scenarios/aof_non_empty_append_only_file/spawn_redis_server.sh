#!/bin/sh
# Buggy append-only scenario: creates AOF dir + manifest like a correct server, but the incr
# file is non-empty with intentionally broken RESP for SET foo 400 (truncated; missing final CRLF).
exec python3 -u - "$@" <<'PY'
import argparse
import os
import socket
import sys


def main() -> None:
    parser = argparse.ArgumentParser(add_help=False)
    parser.add_argument("--port", default="6379")
    parser.add_argument("--dir", default=None)
    parser.add_argument("--appendonly", default=None)
    parser.add_argument("--appenddirname", default=None)
    parser.add_argument("--appendfilename", default=None)
    args, _unknown = parser.parse_known_args()

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
        # Valid manifest line so manifest assertion path is satisfied; broken incr for decode errors.
        with open(manifest_path, "w", encoding="utf-8") as mf:
            mf.write(f"file {incr_basename} seq 1 type i\n")
        # Intended: *3 ... SET ... foo ... 400 — broken by omitting \r\n after bulk payload "400"
        broken_set = b"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n400"
        with open(incr_path, "wb") as af:
            af.write(broken_set)
        sys.stderr.write(
            f"Created {append_dir}, broken RESP in {incr_basename}, {manifest_basename} ok\n"
        )
        sys.stderr.flush()

    port = int(args.port)
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(("0.0.0.0", port))
    sock.listen()
    sys.stderr.write(f"Listening on 0.0.0.0:{port}\n")
    sys.stderr.flush()

    while True:
        conn, _ = sock.accept()
        try:
            conn.recv(4096)
        except OSError:
            pass
        finally:
            conn.close()


if __name__ == "__main__":
    main()
PY
