# blog_aggregator (gator)
boot.dev project for learning SQL+Go

## installation instructions 

* install postgres
* install goose
* install sqlc
* go install

## usage
* create config:
```
echo '{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable","user":"me"}' > ~/.gatorconfig.json
```

* create database:
```bash
psql --version
CREATE DATABASE gator;
cd sql/schema
goose postgres postgres://postgres:postgres@localhost:5432/gator up
cd ../..
```

* run gator
```
go run . register me
go run . addfeed "Lanes Blog" "https://www.wagslane.dev/index.xml"
go run . addfeed "Hacker News RSS" "https://hnrss.org/newest"
go run . agg 5s
go run . browse 5
```

