# protorepo
Monorepo template for TCorp Service Definitions using Protocol Buffers. 


A monorepo is a central repository that contains all proto files for the services within an organization.
Within TCorp, our monorepo is also responsible for generating the client-side libraries from these proto files.
This workflow was heavily inspired by [a blog post by Namely](https://medium.com/namely-labs/how-we-build-grpc-services-at-namely-52a3ae9e7c35).

This repository is for educational purposes. Our actual monorepo is hosted privately.

## Get started
Simply create a directory with your service name (`duck_typed`). 
Within this directory, drop your .proto files and define a [.proto.yaml](greeter/.proto.yaml). 
The protorepo executor should have access to any repositories in the [.proto.yaml](greeter/.proto.yaml), so make sure to create it too.
As a convention, the repository name should have the format of "{service}-proto-{language}".

## Supported languages
Currently the only language that is supported is "go".

## How this repository works
Whenever PR is merged into this repository, [a script](main.go) is executed.

This script will traverse all directories with a [.proto.yaml](greeter/.proto.yaml) file defined. For each language
in the yaml, it will then update this repository in case the generated code has changed.


## How to import (in gitlab)
Make sure that go knows that any gitlab.com project comes from a private git (and can thus not use its  cache) and setup git auth:
```bash
GOPRIVATE=gitlab.com/<org>/*
git config --global url."https://${GIT_USER}:${GIT_PASSWORD}@gitlab.com" .insteadOf "https://gitlab.com"
```

** Only read this if your gitlab uses subgroups**

Golang acts up if the path to the repository uses multiple "/"s. To fix this, we can add .git to the import.

To keep our imports clean we'll have to replace this in the `go.mod` file.

```
module github.com/toonsevrin/example

go 1.14

require (
	gitlab.com/tcorp/protos/greeter-proto-go v0.0.0-20200331155046-028b3003f2e6 // indirect
)

replace gitlab.com/tcorp/protos/greeter-proto-go => gitlab.com/tcorp-k8s/protos/greeter-proto-go.git v0.0.0-20200331155046-028b3003f2e6
```

