# Requires git to have access to all relevant repositories!

deploy: ensure-dependencies
	go run main.go

ensure-dependencies:
	rm -rf /tmp/protoc && mkdir -p /tmp/protoc
	echo 6d0f18cd84b918c7b3edd0203e75569e0c8caecb1367bbbe409b45e28514f5be  - > /tmp/protoc/protosum.txt
	curl -sfL https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip | tee  /tmp/protoc/protoc.zip | sha256sum -c /tmp/protoc/protosum.txt
	unzip  /tmp/protoc/protoc.zip -d /tmp/protoc/
	export PATH=$PATH:/tmp/protoc/protoc/bin
	rm -rf /tmp/protoc
	go get -u github.com/golang/protobuf/protoc-gen-go