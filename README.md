# hosts-override [![Build Status](https://travis-ci.org/rcaught/hosts-override.svg?branch=master)](https://travis-ci.org/rcaught/hosts-override)
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
$client = new-object System.Net.WebClient
$client.DownloadFile("https://github.com/rcaught/hosts-override/releases/latest/download/windows.zip","C:\tmp\hosts-override.zip")
# Then unzip the file and run the bin as Administrator
```

## Usage (Mac / Linux)
```
$ # Override myhost.com to resolve to 127.0.0.1
$ sudo hosts-override myhost.com,127.0.0.1
$ # google.com will be resolved into an IP / set of IPs
$ sudo hosts-override myhost.com,google.com
$ # Multiple hosts and values are supported
$ sudo hosts-override myhost.com,127.0.0.1 anotherhost.com,127.0.0.1
```

### Notes
- `sudo` is required as the hosts file is owned by `root`
- On exiting the program with an interupt (CTRL-c), the hosts file is cleaned of appended records
- In the case of an unclean shutdown, the next invocation of `hosts-override` will clear the previous sessions records
- IPv4 and IPv6 addresses supported
