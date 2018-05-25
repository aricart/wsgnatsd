package server

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/nats-io/gnatsd/logger"
	"github.com/nats-io/gnatsd/server"
)

type NatsServer struct {
	Server *server.Server
	Logger *logger.Logger
}

func NewNatsServer() (*NatsServer, error) {
	// Create a FlagSet and sets the usage
	fs := flag.NewFlagSet("nats-server", flag.ExitOnError)
	opts, err := server.ConfigureOptions(fs, getArgs(),
		server.PrintServerAndExit,
		fs.Usage,
		server.PrintTLSHelpAndDie)

	if err != nil {
		server.PrintAndDie(err.Error() + "\n" + usageString)
	}

	// gnatsd will also handle the signal if this is not set
	opts.NoSigs = true

	s := server.New(opts)

	return &NatsServer{Server: s}, nil
}

func (server *NatsServer) Start() error {
	server.Server.ConfigureLogger()

	go server.Server.Start()

	if !server.Server.ReadyForConnections(5 * time.Second) {
		return errors.New("unable to start embedded gnatsd server")
	}

	return nil
}

func (server *NatsServer) Shutdown() {
	if server.Server != nil {
		server.Server.Shutdown()
	}
}

func (server *NatsServer) ServerPort() int {
	return server.Server.Addr().(*net.TCPAddr).Port
}

func (server *NatsServer) HostPort() string {
	return server.Server.Addr().(*net.TCPAddr).String()
}

func (server *NatsServer) GetURL() string {
	return fmt.Sprintf("nats://%s", server.Server.Addr().String())
}

var usageString = "embedded NATS server options can be supplied by following a '--' argument with any gnatsd supported flag."

func getArgs() []string {
	inlineArgs := -1
	for i, a := range os.Args {
		if a == "--" {
			inlineArgs = i
			break
		}
	}

	var args []string
	if inlineArgs != -1 {
		args = os.Args[inlineArgs+1:]
	} else {
		args = append(args, "-a", "127.0.0.1", "-p", "-1")
	}

	return args
}
