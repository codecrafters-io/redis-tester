# Redis Challenge Tester

This is a program that validates your progress on the Redis challenge.

# Requirements for docker image

- `LOGSTREAM_URL`, passed to `logstream`
- User code mounted at `/app``

Usage:

```
docker run -v <path-to-user-app>:/app -e LOGSTREAM_URL=<logstream_url> redis-tester
```
