package redshirt

import (
	"crypto/tls"

	"github.com/pantheon-systems/go-certauth/certutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewTLSConnection will initiate a connection to riker with given address caFile and client .pem
func NewTLSConnection(addr, caFile, tlsFile string) (*grpc.ClientConn, error) {
	cert, err := certutils.LoadKeyCertFiles(tlsFile, tlsFile)
	if err != nil {
		return nil, fmt.Errorf("could not load TLS cert '%s': %s", tlsFile, err.Error())
	}
	caPool, err := certutils.LoadCACertFile(caFile)
	if err != nil {
		return nil, fmt.Fatalf("Could not load CA cert '%s': %s", caFile, err.Error())
	}
	tlsConfig := certutils.NewTLSConfig(certutils.TLSConfigModern)
	tlsConfig.ClientCAs = caPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.Certificates = []tls.Certificate{cert}

	return grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
}
