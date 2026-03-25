#!/bin/sh
# Buggy fm0 implementation: accepts --dir, --appendonly, --appenddirname, --appendfilename
# and listens on --port, but intentionally does NOT create the append-only directory.
exec python3 -u - "$@" <<'PY'
import argparse
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

    # Correct behavior would mkdir(os.path.join(args.dir, args.appenddirname)) when appendonly is yes.
    # Deliberately omitted so filesystem assertion in fm0 fails.

    port = int(args.port)
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(("0.0.0.0", port))
    sock.listen()
    sys.stderr.write(f"Listening on 0.0.0.0:{port} (AOF dir not created)\n")
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