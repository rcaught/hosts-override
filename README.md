# hosts-override [![Build Status](https://travis-ci.org/rcaught/hosts-override.svg?branch=master)](https://travis-ci.org/rcaught/hosts-override)
Override `hosts` file entries for the lifetime of the process

## Installation

##### MacOS (Homebrew via own tap)
```
$ brew install rcaught/hosts-override/hosts-override
```
##### MacOS (manually)
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
- Download https://github.com/rcaught/hosts-override/releases/latest/download/windows.zip
- Unzip the file and run the exe as Administrator

## Usage
### Mac / Linux
- Override myhost.com to resolve to 127.0.0.1
  ```
  $ sudo hosts-override myhost.com,127.0.0.1
  ```
- google.com will be resolved into an IP / set of IPs
  ```
  $ sudo hosts-override myhost.com,google.com
  ```
- Multiple hosts and values are supported
  ```
  $ sudo hosts-override myhost.com,127.0.0.1 anotherhost.com,127.0.0.1
  ```
- Refresh of unresolved hosts (with custom interval)
  ```
  $ sudo hosts-override -r -i=1m myhost.com,127.0.0.1
  ```
### Windows
- Run Command Prompt as Administrator and navigate to the directory containing hosts-override.exe
- ```
  hosts-override.exe myhost.com,127.0.0.1
  ```

### Notes
- `sudo` or running as Administrator is required as the hosts file is owned by `root`
- On exiting the program with an interupt (CTRL-c), the hosts file is cleaned of appended records
- In the case of an unclean shutdown, the next invocation of `hosts-override` will clear the previous sessions records
- IPv4 and IPv6 addresses supported
