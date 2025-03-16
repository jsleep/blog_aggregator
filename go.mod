module github.com/jsleep/blog_aggregator

go 1.23.5

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	internal/config v1.0.0
)

replace internal/config => ./internal/config
