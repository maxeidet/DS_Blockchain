package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"GO_Blockchain/blockchain"
)

func main() {
	apiAddr := flag.String("api", ":8080", "HTTP API address")
	p2pAddr := flag.String("p2p", ":6001", "P2P TCP listen address")
	advertiseAddr := flag.String("advertise", "localhost:6001", "P2P address advertised to peers")
	peersFlag := flag.String("peers", "", "comma-separated peer addresses")
	discoveryPort := flag.Int("discovery-port", 9999, "UDP discovery port")
	flag.Parse()

	params := blockchain.DefaultNetworkParams()
	bc := blockchain.NewBlockchain(params)

	fmt.Println("Network ID:", params.NetworkID)
	fmt.Println("Genesis balances:")
	for addr, amount := range params.GenesisBalances {
		fmt.Printf("%s -> %d\n", addr, amount)
	}
	fmt.Println()

	var peers []string
	if *peersFlag != "" {
		for _, peer := range strings.Split(*peersFlag, ",") {
			peer = strings.TrimSpace(peer)
			if peer != "" {
				peers = append(peers, peer)
			}
		}
	}

	p2pNode := blockchain.NewP2PNode(
		bc,
		*p2pAddr,
		*advertiseAddr,
		peers,
		*discoveryPort,
	)

	if err := p2pNode.Start(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	api := blockchain.NewAPI(bc, p2pNode)
	api.RegisterRoutes(mux)

	fmt.Println("HTTP API running on", *apiAddr)
	fmt.Println("P2P listen address:", *p2pAddr)
	fmt.Println("Advertise address:", *advertiseAddr)
	fmt.Println("Discovery port:", *discoveryPort)
	fmt.Println("Initial peers:", peers)
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("GET  /status")
	fmt.Println("GET  /blocks")
	fmt.Println("GET  /mempool")
	fmt.Println("GET  /balance/{address}")
	fmt.Println("GET  /peers")
	fmt.Println("POST /transactions")
	fmt.Println("POST /mine")
	fmt.Println()
	fmt.Println("Demo addresses:")
	fmt.Println("addr_alice")
	fmt.Println("addr_bob")
	fmt.Println("addr_charlie")
	fmt.Println()

	log.Fatal(http.ListenAndServe(*apiAddr, mux))
}