# Bug Reproduction

## Scenario

Creating a session with an upload key that is not present in the repository receives a nil session with no error, then the application treats the lookup as a successful existing session.

## Command

`go test ./internal/app -count=1 -run '^TestCreateMissingKeyDoesNotReturnNilSession$'`

## Expected failure

The test fails because creation does not return a valid session and the lookup result can lead to a nil pointer dereference.
