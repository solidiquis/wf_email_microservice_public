# Wf Email Microservice

**Starting up the server**:

```
$ ./wf.sh server 

# Alternatives
$ ./wf.sh s
$ ./wf.sh -p 3333
```

**Go protobuf compilation**:

Place your *.proto files in `api/proto/v1` and run the following:

```
./wf.sh cproto file_name # don't include the .proto extension
```

This will create `pkg/api/v1/file_name.pb.go`.

**Ruby protobuf compilation**:

`protoc --proto_path=app/shared/protobuf/ --ruby_out=app/shared/protobuf/modules app/shared/protobuf/email_microservice.proto`
