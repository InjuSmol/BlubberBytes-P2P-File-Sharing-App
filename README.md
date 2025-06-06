# BlubberBytes

**BlubberBytes** This project implements a decentralized file-sharing and HTTP/HTTPS proxy application with a user-friendly
client-side interface. The primary goal is to develop a distributed system that enables seamless peer-to-peer file
sharing and HTTP proxy connectivity among connected nodes. The application has two core functionalities: file
sharing by hash and an HTTP proxy, which distributes and accesses shared content within the network.
Additionally, the project includes features, such as turning nodes into public HTTP gateways for file access and
supporting the ability to check the shared files of random neighbor nodes.

<img width="759" alt="Screenshot 2025-01-21 at 9 59 23 PM" src="https://github.com/user-attachments/assets/55e03f7f-0076-465c-abb5-061c65aef870" />

## Prerequisites

Make sure you have the following installed on your system:

- **Node.js** (v14 or newer): Download from [nodejs.org](https://nodejs.org/).
- **npm** (Node Package Manager): Comes with Node.js.
- **Electron**: You do **not** need to install Electron globally; it will be installed automatically as part of the project dependencies.
- **Go**: Download from [go.dev/dl/](https://go.dev/dl/)

## Getting Started

Follow these steps to set up and run the application locally:

### Step 1: Clone the Repository

First, clone the repository to your local machine:

```bash
git clone https://github.com/injusmol/BlubberBytes-P2P-File-Sharing-App
```

### Step 2: Build Btcd and Btcwallet

Navigate to the `btcd` and `btcwallet` directories and run the following commands:

```bash
cd blubberbytes/server/btcd
go mod tidy
go clean
go build
```

```bash
cd blubberbytes/server/btcwallet
go mod tidy
go clean
go build
```

### Step 3: Run the Server

Navigate to the `server` directory and run the server:

```bash
cd blubberbytes/server
go run .
```

If btcd or btcwallet fails to start, make sure that the btcd and btcwallet processes are not already running and kill them if they are running. Make sure that the server itself is not already running as well.

You can change the `net` variable in `blubberbytes/server/main.go` to connect to a specific network. It is set to the testnet by default.

### Step 4: Set Up the Client

Navigate to the `client` directory and install the required dependencies:

```bash
cd blubberbytes/client
npm install
```

### Step 5: Build and Run the Client

Once the dependencies are installed, build and run the client with:

```bash
npm run build
npm run electron
```

## Running the Application

After completing the steps above, the GUI should open up automatically.

If the server is running, you can go to http://localhost:3001/generate to generate/mine a block to gain coins. It will take some time before the server responds with the generated block.

## Contributers:

Daniel Liang

Jazz Kaur

John Roedel

Qilong Ren 

Sahibjot Bhullar

InjuSmol
