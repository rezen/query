# awwwq
Query for facts related to your site!


## Requirements
- golang 1.11+


```
# Running tests with specific prefix
go test -v ./ -run TestHttp


# Test some code
staticcheck github.com/rezen/query

# Building
go build -o _builds/awwwq cmd/awwwdit/main.go

# ... or using Docker
docker-compose up

# Usage

./awwwq -t http://ahermosilla.com -q 'http > doc > title'
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


