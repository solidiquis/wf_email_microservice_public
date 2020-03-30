# Wefunder Email Microservice


**Setting up a dev environment**:
```
$ cp .env-example .env
$ heroku git:remote -a wefunder-email-microservice
```

If you want to test actual email sending, get a test API key from SendWithUs and add it to .env as `SWU_API_KEY`. 

To deploy your code to Heroku, commit to master and then run:

```
$ git push heroku master
```

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