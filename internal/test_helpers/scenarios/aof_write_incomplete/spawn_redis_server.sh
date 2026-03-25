#!/bin/sh
# Buggy dc8: creates AOF dir + manifest + empty incr; on SET, appends RESP missing the final \r\n
# after the value bulk string so DecodeCommandsFromAppendOnlyFile fails (incomplete input).
exec python3 -u - "$@" <<'PY'
import argparse
import os
import socket
import sys


def parse_set_key_value(buf: bytes):
    """Parse *3 SET <key> <value> (bulk strings). Returns (key, value) or (None, None)."""
    idx = 0
    if idx >= len(buf) or buf[idx : idx + 1] != b"*":
        return None, None
    line_end = buf.find(b"\r\n", idx)
    if line_end < 0:
        return None, None
    try:
        n_args = int(buf[idx + 1 : line_end])
    except ValueError:
        return None, None
    idx = line_end + 2
    if n_args != 3:
        return None, None
    parts = []
    for _ in range(3):
        if idx >= len(buf) or buf[idx : idx + 1] != b"$":
            return None, None
        line_end = buf.find(b"\r\n", idx)
        if line_end < 0:
            return None, None
        try:
            ln = int(buf[idx + 1 : line_end])
        except ValueError:
            return None, None
        idx = line_end + 2
        if idx + ln > len(buf):
            return None, None
        body = buf[idx : idx + ln]
        idx += ln
        if idx + 2 > len(buf) or buf[idx : idx + 2] != b"\r\n":
            return None, None
        idx += 2
        parts.append(body)
    if parts[0].upper() != b"SET":
        return None, None
    return parts[1], parts[2]


def broken_set_resp(key: bytes, val: bytes) -> bytes:
    # Valid up through value bytes; omit trailing \r\n after the last bulk payload (incomplete RESP).
    return b"".join(
        [
            b"*3\r\n",
            b"$3\r\nSET\r\n",
            (f"${len(key)}\r\n".encode() + key + b"\r\n"),
            (f"${len(val)}\r\n".encode() + val),
        ]
    )


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
            f"Created {append_dir}, empty {incr_basename}, {manifest_basename} ok\n"
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
            data = conn.recv(65536)
            if data and incr_path is not None:
                key, val = parse_set_key_value(data)
                if key is not None and val is not None:
                    with open(incr_path, "ab") as af:
                        af.write(broken_set_resp(key, val))
            conn.sendall(b"+OK\r\n")
        except OSError:
            pass
        finally:
            conn.close()


if __name__ == "__main__":
    main()
PY
