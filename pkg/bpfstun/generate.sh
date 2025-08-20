#!/bin/sh

# Use clilium to generate eBPF go
# https://github.com/cilium/ebpf/tree/main/cmd/bpf2go
BPF2GO=github.com/cilium/ebpf/cmd/bpf2go

# Default will be bpfel, bpfeb
TARGETS="-target bpfel"

# External libraries to use
IMPORTS=

# List of types to generate go decleration
TYPES=

# List of test types to generate go decleration
TEST_TYPES=

BPF2GO_CFLAGS="-O2 -g -Wall -Werror"

INCLUDES="-I/usr/include/x86_64-linux-gnu -I/usr/include/aarch64-linux-gnu"

# To display possible options
# go run $BPF2GO -h

go run $BPF2GO $TARGETS $IMPORTS $TYPES stun stun.c -- $BPF2GO_CFLAGS $INCLUDES &&
go run $BPF2GO $TARGETS $IMPORTS $TEST_TYPES stuntest stun.c -- $BPF2GO_CFLAGS $INCLUDES -DTEST=1
