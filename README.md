# ClawIO Server

# Usage
```
clawio -h
Usage of clawio:
  -config="/etc/sysconfig/clawio.conf": Configuration file location
  -pidfile="/var/run/clawio.pid": PID file location
```
# Boot sequence
1. Parse command line arguments
2. Create PID file
3. Load configuration
4. Start HTTP/HTTPS server to listen to requests
5. Listen to OS signals to control the daemon

# OS signals
The daemon is controlled by sending OS signals to the daemon's process.

- Reload configuration.
```
kill -SIGHUP XXXX
```
- Hard shutdown of the server (not recommended, aborts ongoing requests abruptly)
```
kill -SIGTERM XXXX
kill -SIGINT  XXXX
```
- Graceful shutdown of the server (recommended). Stops accepting new requests and ongoing requests have the opportunity to finish within the timeoout specified in the configuration
```
kill -SIGQUIT XXXX
```

*XXXX* is the daemon's process ID
