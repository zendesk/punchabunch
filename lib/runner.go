package lib

import (
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func handleConn(clientConn net.Conn, sshClient *ssh.Client, proxyAddr string, remoteAddr string, verbose bool) {
	if verbose {
		log.Printf("Requesting SSH server (%s) connect to %s...", proxyAddr, remoteAddr)
	}
	remoteConn, err := sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("Dial: %s: %v", remoteAddr, err)
		return
	}
	if verbose {
		log.Printf("Connection to %s via %s established.", remoteAddr, proxyAddr)
	}

	remoteClosed := make(chan struct{})
	go func() {
		_, err := io.Copy(clientConn, remoteConn)
		if err != nil {
			log.Printf("%s: %v", remoteAddr, err)
		}
		remoteConn.Close()
		clientConn.Close()
		close(remoteClosed)
	}()

	localClosed := make(chan struct{})
	go func() {
		_, err := io.Copy(remoteConn, clientConn)
		if err != nil {
			log.Printf("%s: %v", remoteAddr, err)
		}
		clientConn.Close()
		remoteConn.Close()
		close(localClosed)
	}()

	for remoteClosed != nil || localClosed != nil {
		select {
		case <-remoteClosed:
			remoteClosed = nil
		case <-localClosed:
			localClosed = nil
		}
	}

	if verbose {
		log.Printf("Connection to %s via %s closed.", remoteAddr, proxyAddr)
	}
}

func Run(c Config) {
	// Sanity check
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	if len(sockPath) == 0 {
		log.Fatalf("Environment variable SSH_AUTH_SOCK not defined.  Is an SSH agent running?")
	}
	user := os.Getenv("LOGNAME")
	if len(user) == 0 {
		log.Fatalf("Environment variable LOGNAME not defined.  Set it to the login name you want to connect to the SSH server with.")
	}

	// Connect to agent socket
	sockFile, err := net.Dial("unix", sockPath)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}
	agentClient := agent.NewClient(sockFile)

	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	connectionsMade := 0

	for proxyAddr, connections := range c.connections {
		go func(proxyAddr string, connections []AddressPair) {
			if c.Verbose {
				log.Println("Connecting to " + proxyAddr + " ...")
			}
			sshClient, err := ssh.Dial("tcp", proxyAddr, clientConfig)
			if err != nil {
				log.Fatalf("unable to connect: %v", err)
			}
			if c.Verbose {
				log.Println("Connection established to " + proxyAddr)
			}

			if c.Ready != nil {
				// Once all the connections have been established, we can
				// dispatch a readiness signal
				connectionsMade++
				if connectionsMade == len(c.connections) {
					c.Ready <- true
				}
			}

			for _, connection := range connections {
				if c.Verbose {
					log.Printf("Forwarding connections from %s to %s", connection.ListenAddr, connection.RemoteAddr)
				}
				listener, err := net.Listen("tcp", connection.ListenAddr)
				if err != nil {
					log.Fatalf("listen: %v", err)
				}

				go func(l net.Listener, remoteAddr string) {
					for {
						conn, err := l.Accept()
						if err != nil {
							log.Fatalf("Accept: %v", err)
						}
						if c.Verbose {
							log.Printf("Received connection from %s", conn.RemoteAddr().String())
						}
						go handleConn(conn, sshClient, proxyAddr, remoteAddr, c.Verbose)
					}
				}(listener, connection.RemoteAddr)
			}
		}(proxyAddr, connections)
	}
}
