package core

import (
	"log"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var natsConnInst *nats.Conn
var natsServerInst *server.Server
var subDlsClient = "dls.client.golang"

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Printf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatal("Exiting, no servers available")
	}))
	return opts
}
func closeNatsIo(p *Core) {
	natsConnInst.Close()
	natsServerInst.Shutdown()
	log.Println("closeNatsConnection OK:")
}

func initNatsIo() {
	nsc := &server.Options{}
	nsc.Debug = true
	nsc.Host = "0.0.0.0"
	nsc.Port = 4222
	// Initialize new server with options
	var err error
	natsServerInst, err = server.NewServer(nsc)

	if err != nil {
		panic(err)
	}
	// Start the server via goroutine
	go natsServerInst.Start()

	// Wait for server to be ready for connections
	if !natsServerInst.ReadyForConnections(4 * time.Second) {
		panic("not ready for connection")
	}
	// Connect to server
	// Connect Options.
	optsClient := []nats.Option{nats.Name("NATS Sample Subscriber")}
	optsClient = setupConnOptions(optsClient)
	natsConnInst, err = nats.Connect(natsServerInst.ClientURL(), optsClient...)

	if err != nil {
		panic(err)
	} else {
		// Flush connection to server, returns when all messages have been processed.
		natsConnInst.Flush()
		log.Println("Connected to nats server:", natsServerInst.ClientURL())
		// FlushTimeout specifies a timeout value as well.
		err := natsConnInst.FlushTimeout(1 * time.Second)
		if err != nil {
			log.Println("All clear!")
		} else {
			log.Println("Flushed timed out!")
		}
	}
	log.Println("[Register] Sub url: ", subDlsClient)
	natsConnInst.Subscribe(subDlsClient, func(m *nats.Msg) {
		log.Printf("Received a message: %s\n", string(m.Data))
		m.Respond([]byte("Hello from golang"))
	})
}
