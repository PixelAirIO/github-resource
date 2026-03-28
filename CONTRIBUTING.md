# Contributing

We use [genqclient](https://github.com/Khan/genqlient) to interact with the
GitHub graphql API. This library generates type-safe functions for us that make
the API calls.

To add new functions, add the graphql query in `github.go`. I usually start by
defining the function and then adding the graphql query, like this:

```go
func (g *githubClient) FuncName(<args here, if any>) (ReturnType, error) {
	_ = `# @genqlient
	query ...
	`
}
```

Then run `go generate ./...` which will update `generated.go`.

The query can be placed anywhere, so you could define it outside any function
first, run `go generate ./...`, then create the function and move it inside the
function. I prefer keeping the query inside the function so it's clear where the
query is being used.

# Testing

You can run tests with `go test -v ./...`.

# Building

```
docker build -t github-resource .
```

You can build and run the tests by doing

```
docker build -t github-resource --target tests .
```
