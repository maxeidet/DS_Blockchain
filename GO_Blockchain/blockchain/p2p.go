package blockchain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type InvItem struct {
	Kind string `json:"kind"` // "tx" eller "block"
	ID   string `json:"id"`
}

type InvMessage struct {
	From  string    `json:"from"`
	Items []InvItem `json:"items"`
}

type GetDataMessage struct {
	Items []InvItem `json:"items"`
}

type HandshakeMessage struct {
	Version       int    `json:"version"`
	NetworkID     string `json:"network_id"`
	NodeID        string `json:"node_id"`
	AdvertiseAddr string `json:"advertise_addr"`
	BestHeight    int    `json:"best_height"`
	BestHash      string `json:"best_hash"`
	GenesisHash   string `json:"genesis_hash"`
	KnownPeers    []string `json:"known_peers"`
}

type GetChainMessage struct{}

type ChainMessage struct {
	Blocks []Block `json:"blocks"`
}

type DiscoveryMessage struct {
	NetworkID     string `json:"network_id"`
	NodeID        string `json:"node_id"`
	AdvertiseAddr string `json:"advertise_addr"`
}

var nodeCounter uint64

type P2PNode struct {
	BC            *Blockchain
	NodeAddr      string
	AdvertiseAddr string
	NodeID        string

	DiscoveryPort int

	Peers map[string]bool
	MaxPeers      int
	MaxKnownPeers int
	mu sync.Mutex

	SeenTxs    map[string]bool
	SeenBlocks map[string]bool
}

func NewP2PNode(
	bc *Blockchain,
	nodeAddr string,
	advertiseAddr string,
	peers []string,
	discoveryPort int,
) *P2PNode {
	peerMap := make(map[string]bool)

	maxPeers := 6
	count := 0
	for _, peer := range peers {
		if peer != "" && peer != advertiseAddr {
			if count >= maxPeers {
				break
			}
			peerMap[peer] = true
			count++
		}
	}

	id := atomic.AddUint64(&nodeCounter, 1)

	node := &P2PNode{
		BC:            bc,
		NodeAddr:      nodeAddr,
		AdvertiseAddr: advertiseAddr,
		NodeID:        fmt.Sprintf("node-%d-%s", id, advertiseAddr),
		DiscoveryPort: discoveryPort,
		MaxPeers:      6,
		MaxKnownPeers: 3,
		Peers:         peerMap,
		SeenTxs:       make(map[string]bool),
		SeenBlocks:    make(map[string]bool),
	}

	node.bootstrapSeenState()
	return node
}

func (n *P2PNode) bootstrapSeenState() {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, block := range n.BC.Blocks {
		n.SeenBlocks[block.Hash] = true
		for _, tx := range block.Transactions {
			n.SeenTxs[tx.ID] = true
		}
	}

	for _, tx := range n.BC.Mempool.GetAll() {
		n.SeenTxs[tx.ID] = true
	}
}

func (n *P2PNode) Start() error {
	ln, err := net.Listen("tcp", n.NodeAddr)
	if err != nil {
		return err
	}

	log.Printf("P2P listening on %s\n", n.NodeAddr)
	log.Printf("Advertising P2P address as %s\n", n.AdvertiseAddr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("accept error:", err)
				continue
			}

			go n.handleConnection(conn)
		}
	}()

	go n.bootstrapHandshakeWithPeers()

	if n.DiscoveryPort > 0 {
		if err := n.StartDiscovery(); err != nil {
			return err
		}
	}

	return nil
}

func (n *P2PNode) StartDiscovery() error {
	if n.DiscoveryPort <= 0 {
		return nil
	}

	go n.discoveryListen()
	go n.discoveryBroadcastLoop()

	log.Printf("UDP discovery enabled on port %d\n", n.DiscoveryPort)
	return nil
}

func (n *P2PNode) discoveryListen() {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: n.DiscoveryPort,
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Printf("discovery listen error: %v\n", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 4096)

	for {
		nRead, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("discovery read error: %v\n", err)
			continue
		}

		var msg DiscoveryMessage
		if err := json.Unmarshal(buf[:nRead], &msg); err != nil {
			log.Printf("invalid discovery payload from %v: %v\n", remoteAddr, err)
			continue
		}

		if msg.NetworkID != n.BC.NetworkID {
			continue
		}

		if msg.NodeID == n.NodeID {
			continue
		}

		if msg.AdvertiseAddr == "" || msg.AdvertiseAddr == n.AdvertiseAddr {
			continue
		}

		if n.addPeer(msg.AdvertiseAddr) {
			log.Printf("discovered new peer %s via UDP\n", msg.AdvertiseAddr)

			go func(peer string) {
				if err := n.sendHandshakeToPeer(peer); err != nil {
					log.Printf("handshake to discovered peer %s failed: %v\n", peer, err)
				}
			}(msg.AdvertiseAddr)
		}
	}
}

func (n *P2PNode) discoveryBroadcastLoop() {
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: n.DiscoveryPort,
	}

	conn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		log.Printf("discovery broadcast dial error: %v\n", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	send := func() {
		msg := DiscoveryMessage{
			NetworkID:     n.BC.NetworkID,
			NodeID:        n.NodeID,
			AdvertiseAddr: n.AdvertiseAddr,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("discovery marshal error: %v\n", err)
			return
		}

		if _, err := conn.Write(data); err != nil {
			log.Printf("discovery broadcast write error: %v\n", err)
		}
	}

	send()

	for range ticker.C {
		send()
	}
}

func (n *P2PNode) addPeer(peer string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if peer == "" || peer == n.AdvertiseAddr {
		return false
	}

	if n.Peers[peer] {
		return false
	}

	if len(n.Peers) >= n.MaxPeers {
		log.Printf("peer limit reached (%d), skipping peer %s\n", n.MaxPeers, peer)
		return false
	}

	n.Peers[peer] = true
	return true
}

func (n *P2PNode) bootstrapHandshakeWithPeers() {
	n.mu.Lock()
	peers := make([]string, 0, len(n.Peers))
	for peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.mu.Unlock()

	for _, peer := range peers {
		go func(peer string) {
			if err := n.sendHandshakeToPeer(peer); err != nil {
				log.Printf("handshake to peer %s failed: %v\n", peer, err)
			}
		}(peer)
	}
}

func (n *P2PNode) sendHandshakeToPeer(peer string) error {
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return err
	}
	defer conn.Close()

	hs := n.buildHandshake()
	if err := writeMessage(conn, "handshake", hs); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Println("invalid handshake response message:", err)
			continue
		}

		switch msg.Type {
		case "handshake":
			var remote HandshakeMessage
			if err := json.Unmarshal(msg.Data, &remote); err != nil {
				log.Println("invalid handshake response payload:", err)
				continue
			}
			n.handleHandshake(remote)

		case "chain":
			var chainMsg ChainMessage
			if err := json.Unmarshal(msg.Data, &chainMsg); err != nil {
				log.Println("invalid chain response payload:", err)
				continue
			}
			n.handleChain(chainMsg)

		case "tx":
			var tx Transaction
			if err := json.Unmarshal(msg.Data, &tx); err != nil {
				log.Println("invalid tx response payload:", err)
				continue
			}
			n.handleTx(tx)

		case "block":
			var block Block
			if err := json.Unmarshal(msg.Data, &block); err != nil {
				log.Println("invalid block response payload:", err)
				continue
			}
			n.handleBlock(block)
		}
	}

	return scanner.Err()
}

func (n *P2PNode) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Println("invalid message:", err)
			continue
		}

		switch msg.Type {
		case "handshake":
			var hs HandshakeMessage
			if err := json.Unmarshal(msg.Data, &hs); err != nil {
				log.Println("invalid handshake payload:", err)
				continue
			}

			if !n.handleHandshake(hs) {
				continue
			}

			if err := writeMessage(conn, "handshake", n.buildHandshake()); err != nil {
				log.Println("failed to send handshake response:", err)
			}

		case "getchain":
			n.handleGetChain(conn)

		case "chain":
			var chainMsg ChainMessage
			if err := json.Unmarshal(msg.Data, &chainMsg); err != nil {
				log.Println("invalid chain payload:", err)
				continue
			}
			n.handleChain(chainMsg)

		case "inv":
			var inv InvMessage
			if err := json.Unmarshal(msg.Data, &inv); err != nil {
				log.Println("invalid inv payload:", err)
				continue
			}
			n.handleInv(inv)

		case "getdata":
			var gd GetDataMessage
			if err := json.Unmarshal(msg.Data, &gd); err != nil {
				log.Println("invalid getdata payload:", err)
				continue
			}
			n.handleGetData(conn, gd)

		case "tx":
			var tx Transaction
			if err := json.Unmarshal(msg.Data, &tx); err != nil {
				log.Println("invalid tx payload:", err)
				continue
			}
			n.handleTx(tx)

		case "block":
			var block Block
			if err := json.Unmarshal(msg.Data, &block); err != nil {
				log.Println("invalid block payload:", err)
				continue
			}
			n.handleBlock(block)

		default:
			log.Println("unknown message type:", msg.Type)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("connection read error:", err)
	}
}

func (n *P2PNode) buildHandshake() HandshakeMessage {
	n.mu.Lock()
	peers := make([]string, 0, len(n.Peers))
	for peer := range n.Peers {
		peers = append(peers, peer)
	}
	maxKnownPeers := n.MaxKnownPeers
	n.mu.Unlock()

	if len(peers) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(peers), func(i, j int) {
			peers[i], peers[j] = peers[j], peers[i]
		})
	}

	if len(peers) > maxKnownPeers {
		peers = peers[:maxKnownPeers]
	}

	return HandshakeMessage{
		Version:       1,
		NetworkID:     n.BC.NetworkID,
		NodeID:        n.NodeID,
		AdvertiseAddr: n.AdvertiseAddr,
		BestHeight:    n.BC.Height(),
		BestHash:      n.BC.BestHash(),
		GenesisHash:   n.BC.GenesisHash(),
		KnownPeers:    peers,
	}
}

func (n *P2PNode) handleHandshake(hs HandshakeMessage) bool {
	if hs.NetworkID != n.BC.NetworkID {
		log.Printf("rejecting peer %s: wrong network id %s\n", hs.AdvertiseAddr, hs.NetworkID)
		return false
	}

	if hs.GenesisHash != n.BC.GenesisHash() {
		log.Printf("rejecting peer %s: genesis hash mismatch\n", hs.AdvertiseAddr)
		return false
	}
	
	for _, peer := range hs.KnownPeers {
		if peer == "" || peer == n.AdvertiseAddr {
			continue
		}

		if n.addPeer(peer) {
			log.Printf("learned new peer %s from %s\n", peer, hs.AdvertiseAddr)

			// handshakea direkt med nya peeren
			go func(p string) {
				if err := n.sendHandshakeToPeer(p); err != nil {
					log.Printf("handshake to learned peer %s failed: %v\n", p, err)
				}
			}(peer)
		}
	}

	if hs.AdvertiseAddr != "" && hs.AdvertiseAddr != n.AdvertiseAddr {
		if n.addPeer(hs.AdvertiseAddr) {
			log.Printf("added peer %s from handshake\n", hs.AdvertiseAddr)
		}
	}

	log.Printf(
		"handshake with peer node_id=%s addr=%s height=%d best_hash=%s\n",
		hs.NodeID,
		hs.AdvertiseAddr,
		hs.BestHeight,
		hs.BestHash,
	)

	if hs.BestHeight > n.BC.Height() && hs.AdvertiseAddr != "" {
		log.Printf("peer is ahead, requesting chain from %s\n", hs.AdvertiseAddr)
		go func(peer string) {
			if err := n.requestChainFromPeer(peer); err != nil {
				log.Printf("getchain request to %s failed: %v\n", peer, err)
			}
		}(hs.AdvertiseAddr)
	}

	return true
}

func (n *P2PNode) requestChainFromPeer(peer string) error {
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := writeMessage(conn, "getchain", GetChainMessage{}); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Println("invalid getchain response message:", err)
			continue
		}

		switch msg.Type {
		case "chain":
			var chainMsg ChainMessage
			if err := json.Unmarshal(msg.Data, &chainMsg); err != nil {
				log.Println("invalid chain response payload:", err)
				continue
			}
			n.handleChain(chainMsg)

		case "tx":
			var tx Transaction
			if err := json.Unmarshal(msg.Data, &tx); err != nil {
				log.Println("invalid tx response payload:", err)
				continue
			}
			n.handleTx(tx)

		case "block":
			var block Block
			if err := json.Unmarshal(msg.Data, &block); err != nil {
				log.Println("invalid block response payload:", err)
				continue
			}
			n.handleBlock(block)
		}
	}

	return scanner.Err()
}

func (n *P2PNode) handleGetChain(conn net.Conn) {
	msg := ChainMessage{
		Blocks: cloneBlocks(n.BC.Blocks),
	}

	if err := writeMessage(conn, "chain", msg); err != nil {
		log.Println("failed to send chain:", err)
	}
}

func (n *P2PNode) handleChain(chainMsg ChainMessage) {
	if len(chainMsg.Blocks) == 0 {
		return
	}

	err := n.BC.ReplaceChainIfValid(chainMsg.Blocks)
	if err != nil {
		log.Printf("ignored replacement chain: %v\n", err)
		return
	}

	n.rebuildSeenState()
	log.Printf("replaced local chain with longer valid chain, new height=%d\n", n.BC.Height())
}

func (n *P2PNode) rebuildSeenState() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.SeenTxs = make(map[string]bool)
	n.SeenBlocks = make(map[string]bool)

	for _, block := range n.BC.Blocks {
		n.SeenBlocks[block.Hash] = true
		for _, tx := range block.Transactions {
			n.SeenTxs[tx.ID] = true
		}
	}

	for _, tx := range n.BC.Mempool.GetAll() {
		n.SeenTxs[tx.ID] = true
	}
}

func (n *P2PNode) handleInv(inv InvMessage) {
	if inv.From == "" || inv.From == n.AdvertiseAddr {
		return
	}

	toRequest := make([]InvItem, 0)

	for _, item := range inv.Items {
		switch item.Kind {
		case "tx":
			if !n.hasSeenTx(item.ID) && !n.hasTxInMempool(item.ID) {
				toRequest = append(toRequest, item)
			}
		case "block":
			if !n.hasSeenBlock(item.ID) && !n.hasBlock(item.ID) {
				toRequest = append(toRequest, item)
			}
		}
	}

	if len(toRequest) == 0 {
		return
	}

	log.Printf("requesting data from %s for %d inventory item(s)\n", inv.From, len(toRequest))
	if err := n.requestDataFromPeer(inv.From, toRequest); err != nil {
		log.Printf("getdata request to %s failed: %v\n", inv.From, err)
	}
}

func (n *P2PNode) requestDataFromPeer(peer string, items []InvItem) error {
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return err
	}
	defer conn.Close()

	gd := GetDataMessage{Items: items}
	if err := writeMessage(conn, "getdata", gd); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Println("invalid response message:", err)
			continue
		}

		switch msg.Type {
		case "tx":
			var tx Transaction
			if err := json.Unmarshal(msg.Data, &tx); err != nil {
				log.Println("invalid tx response payload:", err)
				continue
			}
			n.handleTx(tx)

		case "block":
			var block Block
			if err := json.Unmarshal(msg.Data, &block); err != nil {
				log.Println("invalid block response payload:", err)
				continue
			}
			n.handleBlock(block)
		}
	}

	return scanner.Err()
}

func (n *P2PNode) handleGetData(conn net.Conn, gd GetDataMessage) {
	for _, item := range gd.Items {
		switch item.Kind {
		case "tx":
			tx, ok := n.getTxFromMempool(item.ID)
			if !ok {
				continue
			}

			if err := writeMessage(conn, "tx", tx); err != nil {
				log.Println("failed to send tx:", err)
			}

		case "block":
			block, ok := n.getBlockByHash(item.ID)
			if !ok {
				continue
			}

			if err := writeMessage(conn, "block", block); err != nil {
				log.Println("failed to send block:", err)
			}
		}
	}
}

func (n *P2PNode) handleTx(tx Transaction) {
	log.Printf("incoming tx payload: %s -> %s (%d)\n", tx.From, tx.To, tx.Amount)

	if n.hasSeenTx(tx.ID) || n.hasTxInMempool(tx.ID) {
		log.Printf("ignored already-known tx %s\n", tx.ID)
		return
	}

	if err := n.BC.AddTransactionToMempool(tx); err != nil {
		log.Printf("rejected tx %s: %v\n", tx.ID, err)
		return
	}

	n.markTxSeen(tx.ID)
	log.Printf("accepted tx %s -> %s (%d)\n", tx.From, tx.To, tx.Amount)

	n.BroadcastInv([]InvItem{
		{Kind: "tx", ID: tx.ID},
	})
}

func (n *P2PNode) handleBlock(block Block) {
	if n.hasSeenBlock(block.Hash) || n.hasBlock(block.Hash) {
		log.Printf("ignored already-known block %s\n", block.Hash)
		return
	}

	if err := n.HandleIncomingBlock(block); err != nil {
		log.Println("rejected block:", err)
		return
	}

	n.markBlockSeen(block.Hash)
	for _, tx := range block.Transactions {
		n.markTxSeen(tx.ID)
	}

	log.Printf("accepted block index=%d hash=%s\n", block.Index, block.Hash)

	n.BroadcastInv([]InvItem{
		{Kind: "block", ID: block.Hash},
	})
}

func (n *P2PNode) HandleIncomingBlock(block Block) error {
	latest := n.BC.LatestBlock()

	if block.Index <= latest.Index {
		return fmt.Errorf("block already known or older: %d", block.Index)
	}

	if err := IsBlockValid(block, latest); err != nil {
		return err
	}

	if err := n.BC.ValidateBlockTransactions(block.Transactions); err != nil {
		return err
	}

	n.BC.Blocks = append(n.BC.Blocks, block)

	for _, tx := range block.Transactions {
		if tx.Type == "transfer" {
			n.BC.Mempool.Remove(tx.ID)
		}
	}

	return nil
}

func (n *P2PNode) AnnounceTransaction(tx Transaction) {
	n.markTxSeen(tx.ID)

	n.BroadcastInv([]InvItem{
		{Kind: "tx", ID: tx.ID},
	})
}

func (n *P2PNode) AnnounceBlock(block Block) {
	n.markBlockSeen(block.Hash)
	for _, tx := range block.Transactions {
		n.markTxSeen(tx.ID)
	}

	n.BroadcastInv([]InvItem{
		{Kind: "block", ID: block.Hash},
	})
}

func (n *P2PNode) BroadcastInv(items []InvItem) {
	msg := InvMessage{
		From:  n.AdvertiseAddr,
		Items: items,
	}
	n.broadcast("inv", msg)
}

func (n *P2PNode) broadcast(msgType string, payload any) {
	data, err := marshalWireMessage(msgType, payload)
	if err != nil {
		log.Println("marshal broadcast message error:", err)
		return
	}

	n.mu.Lock()
	peers := make([]string, 0, len(n.Peers))
	for peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.mu.Unlock()

	log.Printf("broadcasting message type=%s to peers=%v\n", msgType, peers)

	for _, peer := range peers {
		go func(peer string) {
			conn, err := net.Dial("tcp", peer)
			if err != nil {
				log.Printf("could not connect to peer %s: %v\n", peer, err)
				return
			}
			defer conn.Close()

			if _, err := conn.Write(data); err != nil {
				log.Printf("could not send to peer %s: %v\n", peer, err)
				return
			}

			log.Printf("sent message type=%s to peer %s\n", msgType, peer)
		}(peer)
	}
}

func marshalWireMessage(msgType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Type: msgType,
		Data: raw,
	}

	out, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	out = append(out, '\n')
	return out, nil
}

func writeMessage(conn net.Conn, msgType string, payload any) error {
	data, err := marshalWireMessage(msgType, payload)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func (n *P2PNode) hasSeenTx(id string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.SeenTxs[id]
}

func (n *P2PNode) markTxSeen(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.SeenTxs[id] = true
}

func (n *P2PNode) hasSeenBlock(hash string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.SeenBlocks[hash]
}

func (n *P2PNode) markBlockSeen(hash string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.SeenBlocks[hash] = true
}

func (n *P2PNode) hasTxInMempool(txID string) bool {
	for _, tx := range n.BC.Mempool.GetAll() {
		if tx.ID == txID {
			return true
		}
	}
	return false
}

func (n *P2PNode) getTxFromMempool(txID string) (Transaction, bool) {
	for _, tx := range n.BC.Mempool.GetAll() {
		if tx.ID == txID {
			return tx, true
		}
	}
	return Transaction{}, false
}

func (n *P2PNode) hasBlock(hash string) bool {
	_, ok := n.getBlockByHash(hash)
	return ok
}

func (n *P2PNode) getBlockByHash(hash string) (Block, bool) {
	for _, block := range n.BC.Blocks {
		if block.Hash == hash {
			return block, true
		}
	}
	return Block{}, false
}