# retirejs
A cli tool for retirejs written in golang - aka a single binary!

**Features**  

- Scan a specific JavaScript file
- Scrape JavaScript assets from a url and check for vulnerable libraries
- Scan a directory for vulnerable libraries

## Usage
```shell
retirejs http://example.com
retirejs ~/vcs/test
```

## Building
```
# Fetch code
go get github.com/rezen/retirejs
git clone https://github.com/rezen/retirejs.git

cd ./retirejs/cmd/retirejs/
go build -o retirejs main.go

# Add binary to your path
cp retirejs /usr/bin/retirejs
```

## Todo
- Report formats
- Check sha1 
- Use headless chrome to scrape versions from DOM
- Receive from stdin pipe
- Web directory listing of assets handling
- *Maybe* npm modules? (npm audit does this but single binary is *muy bien*)
