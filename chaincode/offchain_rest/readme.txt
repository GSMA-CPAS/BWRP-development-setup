
prerequisites:

1) have ldbs running
     - connect to mysql via: mysql -h 127.0.0.1 -p -u root
       (use pw from .env)
2) install counterfeiter
GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter


then: using/testing:
- export PATH=$PATH:$(go env GOPATH)/bin
- go generate ./... 
- go test
