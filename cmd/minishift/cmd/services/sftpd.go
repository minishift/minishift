/*
Copyright (C) 2017 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package services

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync/atomic"
)

const (
	sftpdServerPortFlag = "port"
)

var (
	connectionCount uint64
	sftpdServerPort int

	servicesSftpdCmd = &cobra.Command{
		Use:    "sftpd",
		Short:  "Starts sftp server on host for sshfs based host folders.",
		Long:   `Starts sftp server on host for sshfs based host folders.`,
		Run:    runSftp,
		Hidden: true,
	}

	authorizedKeysMap map[string]bool
)

func init() {
	authorizedKeysMap = make(map[string]bool)
	servicesSftpdCmd.Flags().IntVarP(&sftpdServerPort, sftpdServerPortFlag, "p", 2022, "The server port.")
	ServicesCmd.AddCommand(servicesSftpdCmd)
}

func runSftp(cmd *cobra.Command, args []string) {
	serverConfig := serverConfig()
	port := viper.GetInt(config.ServicesSftpPort.Name)
	if port == 0 {
		port = sftpdServerPort
	}

	// Once a ServerConfig has been configured, connections can be accepted.
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		glog.Fatal("failed to listen for connection", err)
	}
	glog.Infof("listening on %v", listener.Addr())

	serveConnections(listener, serverConfig)
}

func serveConnections(listener net.Listener, serverConfig *ssh.ServerConfig) {
	for {
		nConn, err := listener.Accept()
		if err != nil {
			glog.Fatal("failed to accept incoming connection", err)
		}

		// populate authorized keys for handshake
		populateAuthorizedKeysMap()
		// Before use, a handshake must be performed on the incoming net.Conn.
		_, channels, requests, err := ssh.NewServerConn(nConn, serverConfig)
		if err != nil {
			glog.Fatal("failed to handshake", err)
		}
		glog.Info("SSH server established")

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(requests)

		// Service the incoming Channel channel.
		go func() {
			for newChannel := range channels {
				// Channels have a type, depending on the application level
				// protocol intended. In the case of an SFTP session, this is "subsystem"
				// with a payload string of "<length=4>sftp"
				glog.Infof("Incoming channel: %s", newChannel.ChannelType())
				if newChannel.ChannelType() != "session" {
					newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}
				channel, requests, err := newChannel.Accept()
				if err != nil {
					glog.Fatal("could not accept channel.", err)
				}

				// Sessions have out-of-band requests such as "shell",
				// "pty-req" and "env".  Here we handle only the
				// "subsystem" request.
				go func(in <-chan *ssh.Request) {
					for req := range in {
						glog.Infof("Request: %v", req.Type)
						ok := false
						switch req.Type {
						case "subsystem":
							glog.Infof("Subsystem: %s", req.Payload[4:])
							if string(req.Payload[4:]) == "sftp" {
								ok = true
								atomic.AddUint64(&connectionCount, 1)
							}
						}
						req.Reply(ok, nil)
					}
				}(requests)

				serverOptions := []sftp.ServerOption{
					sftp.WithDebug(os.Stderr),
				}
				server, err := sftp.NewServer(
					channel,
					serverOptions...,
				)
				if err != nil {
					glog.Fatal(err)
				}
				if err := server.Serve(); err == io.EOF {
					server.Close()
					atomic.AddUint64(&connectionCount, ^uint64(0))
					currentCount := atomic.LoadUint64(&connectionCount)
					if currentCount == 0 {
						glog.Info("last sftp client exited.")
						os.Exit(0)
					} else {
						glog.Info("sftp client exited session.")
					}
				} else if err != nil {
					glog.Fatal("sftp server completed with error:", err)
				}
			}
		}()
	}
}

func serverConfig() *ssh.ServerConfig {
	// An SSH server is represented by a ServerConfig, which holds certificate details and handles authentication
	config := ssh.ServerConfig{
		PublicKeyCallback: keyAuth,
	}

	if filehelper.Exists(minishiftConstants.ProfilePrivateKeyPath()) {
		privateKeyBytes, err := ioutil.ReadFile(minishiftConstants.ProfilePrivateKeyPath())
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Unable to read private key: %s", err))
		}

		privateKeySigner, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Unable to parse private key: %s", err))
		}
		config.AddHostKey(privateKeySigner)
	}

	return &config
}

// populateAuthorizedKeysMap populates the authorizedKeysMap map after reading the key from the
// authorized_keys file.
// authorized_keys file is maintained separately for each profile.
func populateAuthorizedKeysMap() {
	authorizedKeysBytes, err := ioutil.ReadFile(minishiftConstants.ProfileAuthorizedKeysPath())
	if err != nil {
		glog.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			glog.Fatal(err)
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}
}

func keyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	// Reference link - https://github.com/golang/crypto/blob/master/ssh/example_test.go
	if authorizedKeysMap[string(key.Marshal())] {
		permissions := ssh.Permissions{
			Extensions: map[string]string{
				"pubkey-fp": ssh.FingerprintSHA256(key),
			},
		}
		return &permissions, nil
	}
	return nil, fmt.Errorf("unknown public key for %q", conn.User())
}
