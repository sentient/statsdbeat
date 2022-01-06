# Statsdbeat

Welcome to Statsdbeat.

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/sentient/statsdbeat`

## Getting Started with Statsdbeat

### Requirements

* [Golang](https://golang.org/dl/) > 1.17

### Init Project
To get running with Statsdbeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Statsdbeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/sentient/statsdbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Statsdbeat run the command below. This will generate a binary
in the same directory with the name statsdbeat.

```
make
```


### Run 

```
./statsdbeat -c statsdbeat.yml
```

#### with Debugger

Log output to console. Run Statsdbeat with debugging output enabled:

```
./statsdbeat -c statsdbeat.yml -e -d "statsdbeat"

```
or everything in debug
```
./statsdbeat -c statsdbeat.yml -e -d "*"
```

### Test

To test Statsdbeat, run the following command:

```
make testsuite
```

Send testdata with 
```
echo -n "accounts.authentication.password.failed:1|c" | nc -u -w0 127.0.0.1 8125
echo -n "accounts.authentication.login.time:320|ms" | nc -u -w0 127.0.0.1 8125
echo -n "accounts.authentication.login.num_users:333|g" | nc -u -w0 127.0.0.1 8125

echo -en "n.s.t.cnt1:1|c\n.s.t.nct2:2|c" | nc -u -w0 127.0.0.1 8125
```


alternatively:
```
go test ./... -v
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```


### Cleanup

To clean  Statsdbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Statsdbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/src/github.com/sentient/statsdbeat
git clone https://github.com/sentient/statsdbeat ${GOPATH}/src/github.com/sentient/statsdbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make package
```

This will fetch and create all images required for the build process. The whole process to finish can take several minutes.


## Other stuff


```
./statsdbeat test output
elasticsearch: http://localhost:9200...
  parse url... OK
  connection...
    parse host... OK
    dns lookup... OK
    addresses: ::1, 127.0.0.1
    dial up... OK
  TLS... WARN secure connection disabled
  talk to server... OK
  version: 7.14.2
```

```
./statsdbeat setup template
ILM policy and write alias loading not enabled.

Index setup finished.
```

```
mage
Targets:
  build               builds the Beat binary.
  buildGoDaemon       builds the go-daemon binary (use crossBuildGoDaemon).
  check               formats code, updates generated content, check for common errors, and checks for any modified files.
  clean               cleans all generated files and build artifacts.
  config              generates both the short/reference/docker configs.
  crossBuild          cross-builds the beat for all target platforms.
  fields              generates a fields.yml for the Beat.
  fmt                 formats source code (.go and .py) and adds license headers.
  golangCrossBuild    build the Beat binary inside of the golang-builder.
  package             packages the Beat for distribution.
  test                runs all available tests
  update              updates the generated files (aka make update).
```



## update security packages

```
go get github.com/containerd/containerd@v1.5.9
go get github.com/opencontainers/image-spec@v1.0.2
```