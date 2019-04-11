# awwwq
How did you get a status code with `curl`? What were the `dig` flags & args to get the nameserver for a domain? Without notes or `cheat` you have to search for that info every time. What if instead there was a basic query language to get that kind of info that didn't require you to memorize all the flags? That would be awwwesome!


```sh
# Get a page title
./awwwq -q 'http > doc > title' -t https://ahermosilla.com 

# Get IP we connected to 
./awwwq -q 'http > ip' -t https://ahermosilla.com 

# Get the words on a page
./awwwq  -q 'http > doc > words' -t https://ahermosilla.com

# Check if a site is clean
./awwwq  -q "http > virustotal" -t http://setforspecialdomain.com

# Check the status code for a resource
./awwwq -q "http .git/config > header > status-code" -t http://ahermosilla.com

# Get whois data
./awwwq -q "domain > whois" -t ahermosilla.com

# Get name servers  
./awwwq -q "domain > ns" -t ahermosilla.com
```

## Requirements
- golang 1.11+


### Development
```
# Running tests with specific prefix
go test -v ./ -run TestHttp


# Test some code
staticcheck github.com/rezen/query

# Building
go build -o _builds/awwwq cmd/awwwdit/main.go

# ... or using Docker
docker-compose up
```

## Notes
- CACHE at the edge!


### Go
- https://brandur.org/go-worker-pool
- https://github.com/asaskevich/govalidator
- https://medium.com/learning-the-go-programming-language/writing-modular-go-programs-with-plugins-ec46381ee1a9



## Backlog
- Multi-target query
  - For example, emails from multiple pages
- Default queue and respond to www
- Script engine for extra bits
- Middleware everywhere!
- Extract urls/emails
- Aggregate facts such as `count()`


