
# Stun eBPF lib

eBPF PoC implementation for a lightweight Session Traversal Utilities for NAT (STUN) server implementation

### Dependencies
- llvm 
```
$ sudo apt-get install llvm
```
- [cillium ebpf2go](https://github.com/cilium/ebpf/tree/main/cmd/bpf2go) 
```
$ go mod download github.com/cilium/ebpf
```

### How to run

generate the bpf objects  
```
$ make generate
go generate ./...
```

build the binary  
```
$ make
mkdir -p bin
go build -ldflags "-X "github.com/sifaserdarozen/stun/stun.Version=dirty-d08bcf6" -X "github.com/sifaserdarozen/stun/stun.BuildDate=2025-08-28T23:36:01"" -o bin ./...
```

run  
```
$ sudo ./bin/stun 
2025/08/28 23:42:56 Using configuration...

```

### eBPF debugging

List the programs attaced to the interfaces  
```
$ sudo bpftool net show
xdp:

tc:
lo(1) tcx/ingress ingress prog_id 84 link_id 4 

flow_dissector:

netfilter:

```

tail logs  
```
$ sudo cat /sys/kernel/debug/tracing/trace_pipe
```

### License
MIT License - see [LICENSE](LICENSE) for full text.