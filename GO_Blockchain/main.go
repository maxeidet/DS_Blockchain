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

	nodeWallet, err := blockchain.NewWallet()
	if err != nil {
		log.Fatal(err)
	}

	faucetWallet, err := blockchain.FaucetWallet()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Network ID:", params.NetworkID)
	fmt.Println("Genesis balances:")
	for addr, amount := range params.GenesisBalances {
		fmt.Printf("%s -> %d\n", addr, amount)
	}
	fmt.Println()

	fmt.Println("Node wallet address:", nodeWallet.Address)
	fmt.Println("Faucet wallet address:", faucetWallet.Address)
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
	api := blockchain.NewAPI(bc, p2pNode, nodeWallet, faucetWallet)
	api.RegisterRoutes(mux)

	// CORS middleware — allows the React frontend (any origin) to reach the API
	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})

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
	fmt.Println("POST /transactions/manual")
	fmt.Println("POST /faucet")
	fmt.Println("POST /mine")
	fmt.Println()

	log.Fatal(http.ListenAndServe(*apiAddr, corsHandler))
}