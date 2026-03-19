#!/bin/sh
# Buggy pb9 implementation: creates AOF dir + empty incr file like a correct server, but leaves
# the manifest file empty so AofManifestFileAssertion in test_aof_create_aof_manifest fails.
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
        # Same naming as test_aof_create_aof_manifest / Redis AOF multi-part
        incr_basename = f"{args.appendfilename}.1.incr.aof"
        manifest_basename = f"{args.appendfilename}.manifest"
        incr_path = os.path.join(append_dir, incr_basename)
        manifest_path = os.path.join(append_dir, manifest_basename)
        open(incr_path, "wb").close()
        # Intentionally empty manifest — missing "file <basename> seq 1 type i" line
        open(manifest_path, "wb").close()
        sys.stderr.write(
            f"Created {append_dir}, empty {incr_basename}, empty {manifest_basename}\n"
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
