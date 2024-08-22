package stun

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	MIN_STUN_LEN             = 20
	ID_LEN                   = 12
	MESAGE_COOKIE            = 554828866 // 0x2112a442
	BINDING_REQUEST          = 1         // 0x0001
	BINDING_SUCCESS_RESPONSE = 257       // 0x0101
	MAPPED_ADDRESS           = 1         // 0x0001
	XOR_MAPPED_ADDRESS       = 32        // 0x0020
	IPV4_ATTR                = 1         // 0x0001
)

const (
	NEW_CONN_BUFF_SIZE = 1000
)

type BindRequest struct {
	Type   uint16
	Len    uint16
	Cookie uint32
	ID     [ID_LEN]byte
}

func (self BindRequest) String() string {
	// return fmt.Sprintf("{type: % length: %d}", self.Type, self.Len, self.Cookie, self.ID)
	return fmt.Sprintf("{type: %#04x, length: %d, Cookie: %#04x, ID: %s}", self.Type, self.Len, self.Cookie, hex.EncodeToString(self.ID[:]))
}

type Attribute struct {
	Type uint16
	Len  uint16
}

type AttributeAddress struct {
	AddrType uint16
	Port     uint16
	Addr     uint32
}

type XoredAddress struct {
	Header Attribute
	Addr   AttributeAddress
}

type MappedAddress struct {
	Header Attribute
	Addr   AttributeAddress
}

func NewMappedAddress(port uint16, ip net.IP) (MappedAddress, error) {
	if ip4 := ip.To4(); ip4 != nil {
		return MappedAddress{
			Header: Attribute{
				Type: MAPPED_ADDRESS,
				Len:  8,
			},
			Addr: AttributeAddress{
				AddrType: IPV4_ATTR,
				Port:     port,
				Addr:     binary.BigEndian.Uint32(ip4),
			},
		}, nil
	}

	return MappedAddress{}, errors.New("Not an Ipv4 address")
}

type SuccessBindingResponse struct {
	BindRequest
	MappedAddress
}

// a server implementation
// https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
// canceling the code
// https://stackoverflow.com/questions/66755407/cancelling-a-net-listener-via-context-in-golang
// context
// https://medium.com/@jamal.kaksouri/the-complete-guide-to-context-in-golang-efficient-concurrency-management-43d722f6eaea

func TcpStart(ctx context.Context, conf Configuration, wg *sync.WaitGroup) {
	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		tcpWg := &sync.WaitGroup{}
		newConns := make(chan net.Conn, NEW_CONN_BUFF_SIZE)

		log.Printf("Starting Stun server, listening port at %d/tcp", conf.tcpPort)
		listenTcpUrl := fmt.Sprintf(":%d", conf.tcpPort)
		tcpServer, err := net.Listen("tcp", listenTcpUrl)
		if err != nil {
			log.Fatal(err)
		}

		defer tcpServer.Close()

		// Make listen connections
		tcpWg.Add(1)
		go func(l net.Listener, newConns chan net.Conn, wg *sync.WaitGroup) {
			defer (*wg).Done()
			for {
				c, err := l.Accept()
				log.Printf("... new connection")
				if err != nil {
					log.Printf("Tcp listen error: %s", err)
					// handle error (and then for example indicate acceptor is down)
					newConns <- nil
					return
				}

				newConns <- c
			}
		}(tcpServer, newConns, tcpWg)

	loop:
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping tcp server ...")
				tcpServer.Close()
				break loop
			case conn := <-newConns:
				if nil == conn {
					log.Println("tcp listener stopped ...")
					close(newConns)
					break loop
				}

				tcpWg.Add(1)
				go func(tcpConn net.Conn, wg *sync.WaitGroup) {
					defer (*wg).Done()

					buf := make([]byte, 10000)
					err := tcpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
					if nil != err {
						log.Fatal(err)
					}

					rlen, err := tcpConn.Read(buf)

					if err != nil {
						if !os.IsTimeout(err) {
							log.Println("error: ", err, " read length: ", rlen)
						}
						return
					}

					addrInTcp := tcpConn.RemoteAddr().(*net.TCPAddr)
					// ipAddr := addrInTcp.AddrPort().Addr()
					port := addrInTcp.Port

					if rlen < MIN_STUN_LEN {
						return
					}

					var req BindRequest
					readBuf := bytes.NewReader(buf)
					//
					err = binary.Read(readBuf, binary.BigEndian, &req)
					if nil != err {
						log.Println(err)
					}

					var res SuccessBindingResponse
					res.BindRequest = req
					res.BindRequest.Type = BINDING_SUCCESS_RESPONSE
					res.BindRequest.Len = 12
					res.MappedAddress, _ = NewMappedAddress(uint16(port), addrInTcp.IP)
					fmt.Println(res.BindRequest)

					writeBuf := new(bytes.Buffer)
					err = binary.Write(writeBuf, binary.BigEndian, res)
					if nil != err {
						log.Println(err)
					}

					// Write back the message over UPD
					wlen, err := tcpConn.Write(writeBuf.Bytes())

					if nil != err {
						log.Println(err)
					}

					fmt.Printf("% x is writen %d is send\n", writeBuf.Bytes(), wlen)

					// Shut down the connection.
					tcpConn.Close()

				}(conn, tcpWg)
			}
		}

		log.Printf("Waiting tcp connections to drain")
		tcpWg.Wait()
		close(newConns)
		log.Printf("Tcp connections... drained")
	}()
}

func UdpStart(ctx context.Context, conf Configuration, wg *sync.WaitGroup) {
	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		log.Printf("Starting Stun server, listening port at %d/udp", conf.udpPort)
		listenUrl := fmt.Sprintf(":%d", conf.udpPort)
		udpServer, err := net.ListenPacket("udp", listenUrl)
		if err != nil {
			log.Fatal(err)
		}

		defer udpServer.Close()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping udp server ...")
				return
			default:
				buf := make([]byte, 10000)
				err := udpServer.SetReadDeadline(time.Now().Add(1 * time.Second))
				if nil != err {
					log.Fatal(err)
				}
				rlen, rAddr, err := udpServer.ReadFrom(buf)

				if err != nil {
					if !os.IsTimeout(err) {
						log.Println("error: ", err, " read length: ", rlen)
					}
					continue
				}

				addrInUdp := rAddr.(*net.UDPAddr)
				// ipAddr := addrInUdp.AddrPort().Addr()
				port := addrInUdp.Port

				if rlen < MIN_STUN_LEN {
					continue
				}

				var req BindRequest
				readBuf := bytes.NewReader(buf)
				//
				err = binary.Read(readBuf, binary.BigEndian, &req)
				if nil != err {
					log.Println(err)
				}

				var res SuccessBindingResponse
				res.BindRequest = req
				res.BindRequest.Type = BINDING_SUCCESS_RESPONSE
				res.BindRequest.Len = 12
				res.MappedAddress, _ = NewMappedAddress(uint16(port), addrInUdp.IP)
				fmt.Println(res.BindRequest)

				writeBuf := new(bytes.Buffer)
				err = binary.Write(writeBuf, binary.BigEndian, res)
				if nil != err {
					log.Println(err)
				}

				// Write back the message over UPD
				udpServer.WriteTo(writeBuf.Bytes(), rAddr)

				fmt.Printf("% x", writeBuf.Bytes())
			}
		}
	}()
}
