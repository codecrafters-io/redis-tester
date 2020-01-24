# Redis Challenge Tester

This is a program that validates your progress on the Redis challenge.

# [TODO] Requirements for binary

- Following environment variables:
  - `APP_DIR` - root of the user's code submission

# User code requirements

- A binary named `spawn_redis_server.sh` that spins up the Redis server.
- A file named `codecrafters.yml`, with the following values: 
  - `debug`
  - `current_stage`

# [DEPRECATED] Requirements for docker image

- `LOGSTREAM_URL`, passed to `logstream`
- User code mounted at `/app``

Usage:

```
docker run -v <path-to-user-app>:/app -e LOGSTREAM_URL=<logstream_url> redis-tester
```
