# consul-wrapper

This program allows to run and monitor a command and regisiter it in consul.
It also create TTL health check to make sure that the status of the service is accurate in consul.
If by any change the command stops working it will be deregistered automatically.

## Usage

```
Usage of ./consul-wrapper:
  -args string
        String with all arguments
  -command string
        Command to run
  -frequency int
        Health Check Frequency (in seconds) (default 30)
  -service string
        Consul Service Name
  -token string
        Consul token used for registration
```

## Example

```
 ./consul-wrapper -service plop -frequency 10 -command consul -args "monitor -log-level trace"
Registering 'plop' in consul
Process is running
...
Process is running
Process exited. Exit code:  0
Deregistering 'plop' from consul
```