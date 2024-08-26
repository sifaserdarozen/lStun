package stun

import (
	"net"
	"testing"
)

var testCases = map[string]struct {
	port             uint16
	IPv4             net.IP
	mappedAddress    MappedAddress
	cookie           uint32
	xorMappedAddress XorMappedAddress
}{
	"172.17.0.1:47746": {
		port: 47746,
		IPv4: net.IPv4(172, 17, 0, 1),
		mappedAddress: MappedAddress{
			Header: Attribute{
				Type: MAPPED_ADDRESS,
				Len:  8,
			},
			Addr: AttributeAddress{
				AddrType: IPV4_ATTR,
				Port:     47746,
				Addr:     2886795265,
			},
		},
		cookie: 0x2112a442,
		xorMappedAddress: XorMappedAddress{
			Header: Attribute{
				Type: XOR_MAPPED_ADDRESS,
				Len:  8,
			},
			Addr: AttributeAddress{
				AddrType: IPV4_ATTR,
				Port:     39824,
				Addr:     2365826115,
			},
		},
	},
	"10.0.4.128:32657": {
		port: 32657,
		IPv4: net.IPv4(10, 0, 4, 128),
		mappedAddress: MappedAddress{
			Header: Attribute{
				Type: MAPPED_ADDRESS,
				Len:  8,
			},
			Addr: AttributeAddress{
				AddrType: IPV4_ATTR,
				Port:     32657,
				Addr:     167773312,
			},
		},
		cookie: 0x2112a442,
		xorMappedAddress: XorMappedAddress{
			Header: Attribute{
				Type: XOR_MAPPED_ADDRESS,
				Len:  8,
			},
			Addr: AttributeAddress{
				AddrType: IPV4_ATTR,
				Port:     24195,
				Addr:     722641090,
			},
		},
	},
}

func TestMappedAddress(t *testing.T) {
	for name, test := range testCases {
		// test := test // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mappedAddress, err := NewMappedAddress(test.port, test.IPv4)

			if nil != err {
				t.Errorf("Could not create mapped address with error: %s", err)
			}
			if mappedAddress != test.mappedAddress {
				t.Errorf("MappedAdddress %v is not same as expected %v", mappedAddress, test.mappedAddress)
			}
		})
	}
}

func TestXorMappedAddress(t *testing.T) {

	for name, test := range testCases {
		// test := test // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			xorMappedAddress, err := NewXorMappedAddress(test.port, test.IPv4, test.cookie)

			if nil != err {
				t.Errorf("Could not create xor mapped address with error: %s", err)
			}
			if xorMappedAddress != test.xorMappedAddress {
				t.Errorf("xorMappedAddress %v is not same as expected %v", xorMappedAddress, test.xorMappedAddress)
			}
		})
	}

}
