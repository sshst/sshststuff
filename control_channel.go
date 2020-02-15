package client

import (
	"context"
	"fmt"
	"github.com/sshst/sshststuff/pb"
	"golang.org/x/crypto/ssh"
	"os"
)

type HostKeyer interface {
	GetPublicKey() ssh.PublicKey
	SetCertificate(cert *ssh.Certificate)
}

type controlChannel struct {
	hk HostKeyer
}

var _ pb.ListenerControlServer = &controlChannel{}

func (c *controlChannel) Kill(ctx context.Context, req *pb.KillRequest) (*pb.KillResponse, error) {
	fmt.Printf("Kill requested by %s\n", req.RequesterId)
	os.Exit(0)
	return &pb.KillResponse{}, nil
}

func (c *controlChannel) GetHostKey(context.Context, *pb.GetHostKeyRequest) (*pb.GetHostKeyResponse, error) {
	publicKey := c.hk.GetPublicKey()
	keyBytes := ssh.MarshalAuthorizedKey(publicKey)
	return &pb.GetHostKeyResponse{Key: keyBytes}, nil
}

func (c *controlChannel) PutSignedHostKey(ctx context.Context, req *pb.PutSignedHostKeyRequest) (*pb.PutSignedHostKeyResponse, error) {
	certPub, _, _, _, _ := ssh.ParseAuthorizedKey(req.Key)
	cert := certPub.(*ssh.Certificate)
	fmt.Printf("Received signed host certificate. Listener ID: %s\n\n", cert.KeyId)
	c.hk.SetCertificate(cert)
	return &pb.PutSignedHostKeyResponse{}, nil
}

func (c *controlChannel) PutUrl(ctx context.Context, req *pb.PutUrlRequest) (*pb.PutUrlResponse, error) {
	fmt.Printf("Connect to this session at %s\n", req.Url)
	return &pb.PutUrlResponse{}, nil
}
