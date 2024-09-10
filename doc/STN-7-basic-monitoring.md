# STN-6: Provide basic build and runtime information in monitoring

## The following build and runtime information will be provided
* BUILD_INFO
* UPTIME_SEC

## Corresponding metrics output
```expfmt
...
# HELP lstun_build_info Information about lstun binary
# TYPE lstun_build_info gauge
lstun_build_info{build_date="2024-09-10T15:34:27",env="test",version="dirty-bd23b98"} 1
# HELP lstun_uptime_sec Information about binary uptime
# TYPE lstun_uptime_sec gauge
lstun_uptime_sec 6
...
```

**git rev-parse --short HEAD**  will provide commit hash in short form. We can install git to docker as well.

```bash
~/workbench/stun$ git rev-parse --short HEAD 2>/dev/null
bd23b98
```

we can enrich the version with cleanness of the build with checking if there are uncommited changes. **--porcelain** option will provide machine readable form and **v1** will be used to fix version so that output will be stable, transparent to changes. **2>/dev/null** can be appended for silent fail

```bash
:~/workbench/stun$ git status --porcelain=v1 2>/dev/null
?? a.b
```

Build information can be get with

```bash
:~/workbench/stun$ date +%Y-%m-%dT%H:%M:%S
2024-09-09T07:35:13
```

In summary the following variables will be used

```bash
LSTN_BUILD_VERSION=$(git rev-parse --short HEAD 2>/dev/null) && \
LSTN_BUILD_DATE=$(date +%Y-%m-%dT%H:%M:%S) && \
if ! [[ -z "`git status --porcelain=v1 2>/dev/null`" ]]; then LSTN_BUILD_VERSION="DIRTY ${LSTN_BUILD_VERSION}"; fi
```

Environment will be set to **local** by default, and will be overriden with environment variable in deployments.

## Local Testing
build & run the container
```bash
make docker-build
docker run -p 8081:8081 --env LSTN_INFO_ENV=test stun
```
do local query
```bash
curl localhost:8081/metrics | grep lstun
...
# HELP lstun_build_info Information about lstun binary
# TYPE lstun_build_info gauge
lstun_build_info{build_date="2024-09-10T15:34:27",env="test",version="dirty-bd23b98"} 1
# HELP lstun_uptime_sec Information about binary uptime
# TYPE lstun_uptime_sec gauge
lstun_uptime_sec 8

```