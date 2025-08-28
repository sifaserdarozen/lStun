package bpfstun

//go:generate ./generate.sh

import (
	"errors"
	"log"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

type Specs struct {
	*ebpf.CollectionSpec
	stunSpecs
}

func loadBpfSpecs() (*Specs, error) {
	rawSpecs, err := loadStun()

	if err != nil {
		return nil, err
	}

	specs := &Specs{CollectionSpec: rawSpecs}

	if err := rawSpecs.Assign(&specs.stunSpecs); err != nil {
		return nil, err
	}

	return specs, nil
}

type objects struct {
	stunPrograms
}

type BpfStun struct {
	objects
	ingress link.Link
}

func (bs *BpfStun) Start() {

	specs, err := loadBpfSpecs()

	if err != nil {
		log.Fatalf("Unable to load eBPF specs %v", err)
	}

	// LogLevel:1 will provide verifier logs for success as well
	opts := ebpf.CollectionOptions{Programs: ebpf.ProgramOptions{LogLevel: 1}}
	if err := specs.LoadAndAssign(&bs.objects, &opts); err != nil {
		var verifierError *ebpf.VerifierError = nil
		if errors.As(err, &verifierError) {
			log.Fatalf("Verifier error: %+v\n", verifierError)
		}
		log.Fatalf("Error loading programs into kernel %v", err)
	} else {
		log.Print("All programs successfully loaded and verified")
		log.Print(bs.Ingress.VerifierLog)
	}

	// get listen interface from config
	ifc, err := net.InterfaceByName("lo")
	if err != nil {
		_ = bs.Close()
		log.Fatalf("Could not get interface: %v", err)
	}

	bs.ingress, err = link.AttachTCX(link.TCXOptions{
		Program:   bs.Ingress,
		Attach:    ebpf.AttachTCXIngress,
		Interface: ifc.Index,
	})
	if err != nil {
		_ = bs.Close()
		log.Fatalf("Error attaching bpf program to interface %v", err)
	}
	log.Println("Attached bpf program...")
}

func (bs *BpfStun) Stop() {
	_ = bs.ingress.Close()
	_ = bs.Close()
}
