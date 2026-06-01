# GO_Blockchain & BlockView Explorer

Welcome to **GO_Blockchain** and **BlockView Explorer**! This project consists of a custom, lightweight blockchain written in Go paired with an interactive, modern web explorer built in React + Vite for real-time monitoring, transaction handling, and security analysis.

The project is divided into two main components:
1. **`GO_Blockchain`**: A distributed Go node that supports Proof-of-Work (PoW) consensus, P2P TCP networking, automatic UDP-based peer discovery, and ECDSA cryptographic wallets.
2. **`Frontend`**: A responsive React interface (**BlockView**) that allows you to connect to and monitor multiple nodes simultaneously, submit transactions, trigger mining, and inspect the block structure.

---

## Table of Contents
1. [Project Overview](#project-overview)
2. [Installation & Quick Start Guide](#installation--quick-start-guide)
   - [Prerequisites](#prerequisites)
   - [Running GO_Blockchain Nodes](#running-go_blockchain-nodes)
   - [Running the Frontend (BlockView Explorer)](#running-the-frontend-blockview-explorer)
3. [Networking & P2P (UDP Discovery)](#networking--p2p-udp-discovery)
4. [HTTP API Reference](#http-api-reference)
5. [Security Lab & Attack Suite](#security-lab--attack-suite)

---

## Project Overview

This project is designed to demonstrate the core concepts of a distributed ledger in a clear, hands-on, and visual manner. Key features include:

* **Proof-of-Work (PoW)**: Block validation is secured by solving a hash puzzle with adjustable difficulty.
* **ECDSA Cryptography**: Transactions are signed using Elliptic Curve Cryptography (secp256k1) to guarantee authenticity and non-repudiation.
* **Mempool**: Pending transactions are kept in a local pool before being grouped and mined into a block.
* **P2P Network**: Nodes communicate over TCP to broadcast newly mined blocks and announced transactions.
* **UDP Peer Discovery**: Running nodes automatically discover peers on the same local network without manual configuration.
* **Built-in Faucet**: A dedicated, pre-funded wallet is provided to easily distribute coins for testing purposes.
* **Security Lab**: An integrated security suite that allows you to simulate five classic blockchain attack vectors (replay attacks, signature corruption, public-key masquerading, overdrafts, and timestamp spoofing) to observe how the node's verification logic prevents them.

---

## Installation & Quick Start Guide

### Prerequisites
Before running the project locally, ensure you have the following installed:
* **Go** (version 1.18 or higher) — [Download Go](https://go.dev/dl/)
* **Node.js** (version 18 or higher) & **npm** — [Download Node.js](https://nodejs.org/)

---

### Running GO_Blockchain Nodes

Open your terminal and navigate to the `GO_Blockchain` directory. You can spin up a single node or run multiple nodes locally to form a peer-to-peer network.

#### Start a Single Node (Node 1):
```bash
cd GO_Blockchain
go run . --api :8080 --p2p :6001 --advertise localhost:6001 --discovery-port 9999
```

This starts a blockchain node that:
* Serves the HTTP API on port **`8080`**.
* Listens for P2P TCP connections on port **`6001`**.
* Advertises its P2P address as `localhost:6001` to other nodes.
* Employs UDP port **`9999`** to listen for and broadcast discovery pings.

#### Go Application Flags:
When launching `go run .`, you can customize the node's behavior with the following options:
* `--api`: Address and port for the HTTP API (default `:8080`).
* `--p2p`: Port for incoming P2P TCP connections (default `:6001`).
* `--advertise`: The P2P address (IP:port) advertised to other nodes (default `localhost:6001`).
* `--peers`: A comma-separated list of static peer addresses to connect to (e.g. `localhost:6001,localhost:6002`), bypassing automatic discovery if preferred.
* `--discovery-port`: The UDP port used by the automatic node discovery service (default `9999`).

---

### Running the Frontend (BlockView Explorer)

Open a new terminal window and navigate to the `Frontend` directory.

1. **Install dependencies**:
   ```bash
   cd Frontend
   npm install
   ```

2. **Start the development server**:
   ```bash
   npm run dev
   ```

3. **Open the Explorer in your browser**:
   Vite will serve the app (typically at [http://localhost:5173](http://localhost:5173)). Open this URL in your web browser.

4. **Connect to your node**:
   * Locate the input field in the top-left sidebar of the BlockView interface.
   * Enter the URL of your running Go node (e.g., `http://localhost:8080`) and click **`+`**.
   * The explorer will establish a connection, showing block history, mempool items, wallet information, and live stats!

---

## Networking & P2P (UDP Discovery)

Thanks to the **automatic UDP discovery** service, multiple nodes started on your computer (or across different devices on the same local network) will instantly find one another and sync up without needing to define static peers.

### Example: Running a local 2-Node P2P Network

Open three terminal windows:

* **Terminal 1 (Node 1)**:
  ```bash
  cd GO_Blockchain
  go run . --api :8080 --p2p :6001 --advertise localhost:6001
  ```

* **Terminal 2 (Node 2)**:
  ```bash
  cd GO_Blockchain
  go run . --api :8081 --p2p :6002 --advertise localhost:6002
  ```

* **Terminal 3 (Frontend)**:
  ```bash
  cd Frontend
  npm run dev
  ```

Once both nodes are running, add both `http://localhost:8080` and `http://localhost:8081` in the BlockView sidebar to monitor them side by side.

* If you submit a transaction to Node 1, it will be propagated to Node 2's mempool via P2P.
* If you mine a block on Node 2, the new block is broadcast to Node 1, which validates it and appends it to its own local copy of the chain.

---

## HTTP API Reference

Each Go node exposes a comprehensive, REST-like HTTP API. You can interact with these endpoints using `curl`, Postman, or our built-in React UI.

All responses are returned as `application/json`.

### 1. Get Node Status
* **Method**: `GET`
* **Path**: `/status`
* **Description**: Returns the node's current height, consensus rules, active mempool size, chain validity state, and node/faucet wallet addresses.
* **Example Response**:
  ```json
  {
    "height": 4,
    "difficulty": 3,
    "mining_reward": 50,
    "mempool_size": 0,
    "is_valid": true,
    "wallet_address": "04a8b...",
    "wallet_nonce": 2,
    "faucet_address": "04c2d..."
  }
  ```

### 2. Get Blocks
* **Method**: `GET`
* **Path**: `/blocks`
* **Description**: Retrieves the complete list of blocks in the blockchain from the genesis block to the tip.
* **Example Response**: An array of block objects, each containing:
  ```json
  [
    {
      "index": 0,
      "timestamp": 1717243912000,
      "transactions": [],
      "prev_hash": "0000000000000000000000000000000000000000000000000000000000000000",
      "hash": "000a12b3c4...",
      "nonce": 42091,
      "difficulty": 3
    }
  ]
  ```

### 3. Get Mempool
* **Method**: `GET`
* **Path**: `/mempool`
* **Description**: Fetches all pending transactions currently in the mempool waiting to be mined.
* **Example Response**:
  ```json
  [
    {
      "id": "8f3b...",
      "from": "04a8b...",
      "to": "04e9c...",
      "amount": 25,
      "timestamp": 1717244015000,
      "type": "transfer",
      "nonce": 1,
      "public_key": "04a8b...",
      "signature": "30450221..."
    }
  ]
  ```

### 4. Get Balance
* **Method**: `GET`
* **Path**: `/balance/{address}`
* **Description**: Returns the calculated coin balance for a specific wallet address based on the current blockchain state.
* **Example Response**:
  ```json
  {
    "address": "04a8b3f...",
    "balance": 175
  }
  ```

### 5. Get Connected Peers
* **Method**: `GET`
* **Path**: `/peers`
* **Description**: Lists all peer nodes currently connected over the TCP network.
* **Example Response**:
  ```json
  [
    "localhost:6002"
  ]
  ```

### 6. Create Standard Transaction
* **Method**: `POST`
* **Path**: `/transactions`
* **Description**: Creates and signs a new transfer transaction **originating from the node's own wallet**. The node automatically queries the next valid nonce, signs the transaction with its private key, adds it to its local mempool, and broadcasts it to the P2P network.
* **Request Body (JSON)**:
  ```json
  {
    "to": "04e9c8a...",
    "amount": 25
  }
  ```
* **Example Response**:
  ```json
  {
    "message": "signed transaction added to mempool",
    "tx": {
      "id": "8f3b...",
      "from": "04a8b...",
      "to": "04e9c8a...",
      "amount": 25,
      "timestamp": 1717244015000,
      "type": "transfer",
      "nonce": 1,
      "public_key": "04a8b...",
      "signature": "30450221..."
    }
  }
  ```

### 7. Create Transaction with Manual Nonce
* **Method**: `POST`
* **Path**: `/transactions/manual`
* **Description**: Similar to `/transactions` but lets you specify a custom `nonce` value manually. This is highly useful for testing validation rules (e.g., verifying that out-of-order or duplicate nonces are rejected).
* **Request Body (JSON)**:
  ```json
  {
    "to": "04e9c8a...",
    "amount": 10,
    "nonce": 5
  }
  ```

### 8. Request Faucet Funds
* **Method**: `POST`
* **Path**: `/faucet`
* **Description**: Requests test coins from the pre-funded Faucet wallet to be transferred to a specific target address.
* **Request Body (JSON)**:
  ```json
  {
    "to": "04e9c8a...",
    "amount": 100
  }
  ```

### 9. Mine a Block
* **Method**: `POST`
* **Path**: `/mine`
* **Description**: Collects all transactions in the mempool, appends a coinbase transaction (mining reward) for the designated miner address, solves the Proof-of-Work puzzle, and broadcasts the completed block to the network.
* **Request Body (JSON)**:
  ```json
  {
    "miner_address": "04a8b3f..."
  }
  ```
* **Example Response**:
  ```json
  {
    "message": "block mined",
    "block": {
      "index": 5,
      "timestamp": 1717244299000,
      "transactions": [...],
      "prev_hash": "000a12b3c4...",
      "hash": "00098f7e2a...",
      "nonce": 99312,
      "difficulty": 3
    }
  }
  ```

### 10. Run Attack Tests (Security Lab)
* **Method**: `POST`
* **Path**: `/attack-test`
* **Description**: Triggers an internal, isolated security suite that executes five classic blockchain attack scenarios in a sandbox, returning detailed verification results showing whether each attack was successfully blocked.
* **Example Response**:
  ```json
  {
    "attacks": [
      {
        "name": "Replay Attack",
        "description": "Submit same signed transaction twice. Second attempt must be rejected.",
        "blocked": true,
        "error": "nonce already used or duplicate tx ID",
        "expected": "Rejected: duplicate tx ID or nonce already used"
      }
    ],
    "summary": {
      "total": 5,
      "blocked": 5,
      "leaked": 0
    }
  }
  ```

---

## Security Lab & Attack Suite

When clicking **"Run Attack Suite"** under the **Security** tab in the BlockView Explorer UI, a request is made to the `/attack-test` endpoint, which simulates the following five attack scenarios on an isolated chain sandbox:

1. **Replay Attack**:
   * *The Attack*: An attacker intercepts a previously sent, valid signed transaction and attempts to broadcast it again in an effort to double-spend.
   * *Protection*: The node keeps track of processed transaction IDs and strictly enforces sequential `nonces` for each active address. Resubmitting an identical transaction or using an outdated nonce is rejected.

2. **Invalid Signature**:
   * *The Attack*: An attacker modifies transaction parameters (like destination or amount) or submits a transaction signed with corrupted signature bytes.
   * *Protection*: The node performs an ECDSA cryptographic signature check against the sender's public key and the hashed transaction contents before admitting it to the mempool.

3. **Wrong Public Key**:
   * *The Attack*: An attacker signs a transaction using their own private key (e.g., Eve's) but attaches a public key and sender address belonging to a high-balance wallet (e.g., Alice's).
   * *Protection*: The node derives the sender address directly from the provided public key. If this derived address does not match the stated `from` address, the transaction is discarded.

4. **Overdraft**:
   * *The Attack*: A user attempts to transfer more coins than their account balance currently holds.
   * *Protection*: Prior to adding a transaction to the mempool or block, the node calculates the historical UTXO/account balance for the sender. If their funds are insufficient, the transaction is blocked.

5. **Future Timestamp**:
   * *The Attack*: A malicious miner attempts to mine a block and sets its timestamp far into the future (e.g., 3 hours ahead) to exploit difficulty adjustments or timestamp rules.
   * *Protection*: Block validation checks the incoming block's timestamp against the node's local system time. If the block's time exceeds a logical safety window (e.g., 2 hours ahead), it is deemed invalid and discarded.

---

Have fun exploring your distributed ledger! If you have any questions or want to scale your network, simply run more nodes following the steps outlined above.