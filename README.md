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
