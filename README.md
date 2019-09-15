# hosts-override
Override `hosts` file entries for the lifetime of the process

## Installation
##### Go
```
$ go get github.com/rcaught/hosts-override/...
```
##### MacOS
```
$ curl -Ls https://github.com/rcaught/hosts-override/releases/latest/download/macos.zip > /tmp/hosts-override.zip
$ unzip /tmp/hosts-override.zip -d /usr/local/bin
```
##### Linux
```
$ curl -Ls https://github.com/rcaught/hosts-override/releases/latest/download/linux.zip > /tmp/hosts-override.zip
$ unzip /tmp/hosts-override.zip -d /usr/local/bin
```
##### Windows
```
$ curl -Ls https://github.com/rcaught/hosts-override/releases/latest/download/windows.zip > /tmp/hosts-override.zip
$ unzip /tmp/hosts-override.zip -d /usr/local/bin
```

## Usage (Mac / Linux)
```
$ sudo hosts-override -0 myhost.com -1 127.0.0.1 # Override myhost.com to point to localhost
$ sudo hosts-override -0 myhost.com -1 google.com # google.com will be resolved into an IP / set of IPs
$ sudo hosts-override -o myhost.com -1 127.0.0.1 -0 anotherhost.com -1 127.0.0.1 # Multiple hosts and values are supported
```

### Notes
- `sudo` is required as the hosts file is owned by `root`
- On exiting the program with an interupt (CTRL-c), the hosts file is cleaned of appended records
- In the case of an unclean shutdown, the next invocation of `hosts-override` will clear the previous sessions records
