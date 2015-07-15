# ClawIO Daemon
The ClawIO Daemon is the core part of the ClawIO project.
It provides an HTTP/HTTPS server able to handle thousands of requests in a very high scalable way.
It offers a full featured API that allows synchronisation and sharing over modern high-performance and multi PB storages.
 
# Usage
```
clawiod -h
Usage of clawiod:
  -c="": Configuration file
  -p="": PID file
  -pc=false: Prints the default configuration file
```
To run the daemon you need to specify where the PID and configuration files are.
If you don't have a configuration file yet,  you can create a default one running ```clawiod -pc```

# Boot sequence
The following steps are done by the daemon to boot:
1. Parse command line arguments
2. Create PID file
3. Load configuration
4. Connect to syslog daemon
5. Load authorization providers (file, sql, LDAP, SSO, kbr5...)
6. Load storage providers (local, eos, s3...)
7. Load APIs
8. Start HTTP/HTTPS server to listen to requests
9. Listen to OS signals to control the daemon

# How to control the daemon
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
- Graceful shutdown of the server (recommended). Stops accepting new requests and ongoing requests have the opportunity to finish within the timeout specified in the configuration
```
kill -SIGQUIT XXXX
```

*XXXX* is the daemon's process ID

# License
See COPYING file