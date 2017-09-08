package riker

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"log"

	"github.com/pantheon-systems/riker/pkg/botpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// TODO: move this into config infrastructure. client certs must contain one of these OUs to connect
var allowedOUs = []string{
	"riker-redshirt",
	"riker-server",
	"riker",
}

// clientCertKey is the context key where the client mTLS cert (pkix.Name) is stored
type clientCertKey struct{}

type rikerError string

func (e rikerError) Error() string {
	return string(e)
}

const (
	// ErrorAlreadyRegistered is returned when a registration attempt is made for an
	// already existing command and the Capability deosn't specify forced registration
	ErrorAlreadyRegistered = rikerError("Already registered")

	// ErrorNotRegistered is returned when a client making a request request is not registered.
	ErrorNotRegistered = rikerError("Not registered")

	// ErrorInternalServer is a Generic failure on the server side while trying to process a client (usually a registration)
	ErrorInternalServer = rikerError("Unable to complete request")

	// RedshirtBacklogSize is the redshirt message queue buffer size. how many messages we will backlog for a registerd command
	RedshirtBacklogSize = uint32(20)
)

type Bot interface {
	Run()
	ChatSend(msg, channel string)
	botpb.RikerServer
}

// intersectArrays returns true when there is at least one element in common between the two arrays
func intersectArrays(orig, tgt []string) bool {
	for _, i := range orig {
		for _, x := range tgt {
			if i == x {
				return true
			}
		}
	}
	return false
}

// tlsConnStateFromContext extracts the client (peer) tls connectionState from a context
func tlsConnStateFromContext(ctx context.Context) (*tls.ConnectionState, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.PermissionDenied, "Permission denied: no peer info")
	}
	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, grpc.Errorf(codes.PermissionDenied, "Permission denied: peer didn't not present valid peer certificate")
	}
	return &tlsInfo.State, nil
}

// getCertificateSubject extracts the subject from a verified client certificate
func getCertificateSubject(tlsState *tls.ConnectionState) (pkix.Name, error) {
	if tlsState == nil {
		return pkix.Name{}, grpc.Errorf(codes.PermissionDenied, "request is not using TLS")
	}
	if len(tlsState.PeerCertificates) == 0 {
		return pkix.Name{}, grpc.Errorf(codes.PermissionDenied, "no client certificates in request")
	}
	if len(tlsState.VerifiedChains) == 0 {
		return pkix.Name{}, grpc.Errorf(codes.PermissionDenied, "no verified chains for remote certificate")
	}

	return tlsState.VerifiedChains[0][0].Subject, nil
}

// authOU extracts the client mTLS cert from a context and verifies the client cert OU is in the list of
// allowedOUs. It returns an error if the client is not permitted.
func AuthOU(ctx context.Context) (context.Context, error) {
	tlsState, err := tlsConnStateFromContext(ctx)
	if err != nil {
		return nil, err
	}

	clientCert, err := getCertificateSubject(tlsState)
	if err != nil {
		return nil, err
	}
	log.Printf("authOU: client CN=%s, OU=%s", clientCert.CommonName, clientCert.OrganizationalUnit)

	if intersectArrays(clientCert.OrganizationalUnit, allowedOUs) {
		return context.WithValue(ctx, clientCertKey{}, clientCert), nil
	}
	log.Printf("authOU: client cert failed authentication. CN=%s, OU=%s", clientCert.CommonName, clientCert.OrganizationalUnit)
	return nil, grpc.Errorf(codes.PermissionDenied, "client cert OU '%s' is not allowed", clientCert.OrganizationalUnit)
}
