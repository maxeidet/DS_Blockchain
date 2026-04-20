package main

import (
	"ds-project/Blockchain"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var bc = blockchain.NewBlockchain()

func getChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(bc.Chain)
}

func mineBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var body struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Data == "" {
		http.Error(w, `{"error":"provide a JSON body with a 'data' field"}`, http.StatusBadRequest)
		return
	}

	newBlock := bc.AddBlock(body.Data)
	json.NewEncoder(w).Encode(newBlock)
}

func registerNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var body struct {
		Nodes []string `json:"nodes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Nodes) == 0 {
		http.Error(w, `{"error":"Please supply a valid list of nodes in a JSON body"}`, http.StatusBadRequest)
		return
	}

	for _, node := range body.Nodes {
		bc.RegisterNode(node)
	}

	response := map[string]interface{}{
		"message":     "New nodes have been added",
		"total_nodes": len(bc.Nodes),
	}
	json.NewEncoder(w).Encode(response)
}

func resolveConflicts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	replaced := bc.ResolveConflicts()

	response := map[string]interface{}{
		"message": "Our chain is authoritative",
		"chain":   bc.Chain,
	}
	if replaced {
		response["message"] = "Our chain was replaced"
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	port := flag.String("port", "8080", "Port to run the HTTP server on")
	flag.Parse()

	http.HandleFunc("/chain", getChain)
	http.HandleFunc("/mine", mineBlock)
	http.HandleFunc("/nodes/register", registerNodes)
	http.HandleFunc("/nodes/resolve", resolveConflicts)

	serverAddr := fmt.Sprintf(":%s", *port)
	log.Printf("Blockchain node running on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
