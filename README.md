# <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original-wordmark.svg" width="40" height="40"/> Shortify · Distributed URL Shortener

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
  <img src="https://img.shields.io/badge/gRPC-✓-00ADD8?style=for-the-badge&logo=grpc&logoColor=white" alt="gRPC"/>
  <img src="https://img.shields.io/badge/Raft-Consensus-FF6C37?style=for-the-badge&logo=ethereum&logoColor=white" alt="Raft Consensus"/>
  <img src="https://img.shields.io/badge/Docker-✓-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"/>
  <img src="https://img.shields.io/badge/Render-Deployed-46E3B7?style=for-the-badge&logo=render&logoColor=white" alt="Render"/>
  <img src="https://img.shields.io/badge/License-MIT-FFD700?style=for-the-badge&logo=opensourceinitiative&logoColor=white" alt="License"/>
</p>

---

## 🚀 Live Demo

**🔗 Try it now:** [https://url-shortener-wtor.onrender.com](https://url-shortener-wtor.onrender.com)

---

## 📸 Screenshot

![Shortify UI](url.png)

*The clean, dark-themed web interface showing URL shortening in action.*

---

## 🚀 Overview

**Shortify** is a **fault-tolerant, distributed URL shortening service** built with Go, Raft Consensus, and gRPC. It can run as a single node or scale to a 3-node cluster that survives server failures without losing data.

---

## 🏗️ System Architecture

### High-Level Architecture
┌──────────────────────────────────────────────────────┐
│ CLIENT │
│ (Browser / gRPC Client) │
└─────────────────────────┬────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────┐
│ REVERSE PROXY │
│ (Render.com Load Balancer) │
└─────────────────────────┬────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────┐
│ HTTP / gRPC SERVER │
│ (Port 8080 / 50051) │
│ │
│ ┌─────────────────────────────────────────────┐ │
│ │ ROUTING LAYER │ │
│ │ GET / → Web UI (HTML) │ │
│ │ POST /shorten → API endpoint │ │
│ │ GET /{code} → Redirect handler │ │
│ └─────────────────────────────────────────────┘ │
└─────────────────────────┬────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────┐
│ RAFT CONSENSUS LAYER │
│ │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │ LEADER │ │ FOLLOWER │ │ FOLLOWER │ │
│ │ Node 1 │◄─┤ Node 2 │◄─┤ Node 3 │ │
│ │ (Elected) │ │ (Replica) │ │ (Replica) │ │
│ └──────┬──────┘ └─────────────┘ └─────────────┘ │
│ │ │
│ ▼ │
│ ┌─────────────────────────────────────────────┐ │
│ │ LOG REPLICATION & CONSENSUS │ │
│ │ All writes go through Leader │ │
│ │ Leader replicates to Followers │ │
│ │ Majority quorum required (2/3) │ │
│ └─────────────────────────────────────────────┘ │
└─────────────────────────┬────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────┐
│ FSM (Finite State Machine) │
│ │
│ ┌─────────────────────────────────────────────┐ │
│ │ IN-MEMORY KEY-VALUE STORE │ │
│ │ │ │
│ │ ┌─────────────────────────────────────┐ │ │
│ │ │ short_code → long_url │ │ │
│ │ │ "le.com" → "https://google.com"│ │ │
│ │ │ "on.com" → "https://amazon.com"│ │ │
│ │ │ "b.com" → "https://github.com"│ │ │
│ │ └─────────────────────────────────────┘ │ │
│ │ │ │
│ │ 🔒 Thread-safe with sync.RWMutex │ │
│ │ ✅ Multiple readers, single writer │ │
│ └─────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘

text

### Request Flow: URL Shortening
User enters URL → 2. Browser sends POST /shorten
│
▼

Server checks Raft state → Is this node the Leader?
│
┌─────┴─────┐
│ │
YES NO
│ │
▼ ▼

Generate Return Redirect to
short code "Not the Leader
(6 chars) Leader"
│
▼

Raft applies
the command
to the FSM
│
▼

Data is saved
and replicated
to Followers
│
▼

Return short
code to client

text

### Request Flow: URL Resolution (Redirect)
User clicks short link → https://shortify.com/le.com
│
▼

Server extracts code → "le.com"
│
▼

FSM resolves code → "https://google.com"
│
┌─────┴─────┐
│ │
Found Not Found
│ │
▼ ▼

HTTP 302 Return
Redirect to 404 Not
google.com Found

text

### Data Flow: Raft Consensus
┌──────────────────────────────────────────────────────┐
│ CLIENT │
└─────────────────────────┬────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────┐
│ LEADER │
│ ┌─────────────────────────────────────────────┐ │
│ │ 1. Receives write request │ │
│ │ 2. Appends to local log │ │
│ │ 3. Replicates log to followers │ │
│ │ 4. Waits for majority ACK │ │
│ │ 5. Commits entry │ │
│ │ 6. Applies to FSM │ │
│ │ 7. Responds to client │ │
│ └─────────────────────────────────────────────┘ │
└───────────┬───────────────────────┬─────────────────┘
│ │
▼ ▼
┌─────────────────────┐ ┌─────────────────────┐
│ FOLLOWER 1 │ │ FOLLOWER 2 │
│ ┌───────────────┐ │ │ ┌───────────────┐ │
│ │ 1. Append log │ │ │ │ 1. Append log │ │
│ │ 2. ACK to │ │ │ │ 2. ACK to │ │
│ │ Leader │ │ │ │ Leader │ │
│ │ 3. Commit │ │ │ │ 3. Commit │ │
│ │ 4. Apply FSM │ │ │ │ 4. Apply FSM │ │
│ └───────────────┘ │ │ └───────────────┘ │
└─────────────────────┘ └─────────────────────┘

text

---

## ✨ Features

| Feature | Description |
| :--- | :--- |
| ⚡ **Instant Shortening** | Paste any long URL and get a short link instantly. |
| 🔄 **Redirect** | Click the short link and get redirected to the original URL. |
| 📋 **Copy to Clipboard** | One-click copy with "Copied!" feedback. |
| 📊 **Character Savings** | Shows how many characters you saved. |
| 🌐 **gRPC API** | Programmatic access with Protocol Buffers. |
| 🛡️ **Fault-Tolerant** | Built with Raft consensus – survives node failures. |
| 🐳 **Containerized** | Runs anywhere with Docker. |
| 🌍 **Cloud Deployed** | Live on Render.com with auto-https. |

---

## 🛠️ Tech Stack

| Component | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go 1.21+ | High-performance backend |
| **RPC** | gRPC + Protocol Buffers | Fast, typed API communication |
| **Consensus** | Hashicorp Raft | Leader election & log replication |
| **Networking** | TCP / HTTP/2 | Reliable communication |
| **Container** | Docker + Docker Compose | Consistent deployment |
| **Deployment** | Render.com | Cloud hosting with free tier |
| **Frontend** | HTML + CSS + JavaScript | Interactive web UI |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker (optional)
- `protoc` (for gRPC code generation)

### Run Locally

```bash
# 1. Clone the repository
git clone https://github.com/SyedSaaduddin1101/URL-Shortener.git
cd URL-Shortener

# 2. Generate gRPC code
protoc --go_out=. --go-grpc_out=. --go_opt=module=distributed-url-shortener --go-grpc_opt=module=distributed-url-shortener proto/kv.proto

# 3. Download dependencies
go mod tidy

# 4. Run the server
go run cmd/server/main.go

# 5. Open your browser
open http://localhost:8080
Run with Docker
bash
# Build and run with Docker Compose
docker compose up --build
📡 API Endpoints
HTTP Endpoints
Endpoint	Method	Description
/	GET	Web UI
/shorten	POST	Shorten a URL
/{code}	GET	Redirect to original URL
gRPC Endpoints (Port 50051)
Service	Method	Description
URLShortener	Shorten	Create a short link
URLShortener	Resolve	Resolve a short link
🧪 Testing
bash
# Run the gRPC client
go run client.go

# Expected output:
# 🔗 Shortened: https://google.com -> Code: le.com
# 🔍 Resolved: Code le.com -> URL: https://google.com
📁 Project Structure
text
URL-Shortener/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── fsm/
│   │   └── fsm.go           # Raft FSM (Database)
│   └── server/
│       └── grpc.go          # gRPC Handlers
├── proto/
│   └── kv.proto             # API Contract
├── static/
│   └── index.html           # Web UI
├── client.go                # gRPC Client
├── Dockerfile               # Container Build
├── render.yaml              # Render Deployment
├── go.mod
└── README.md
🌐 Deployment
This project is deployed on Render.com:

Push code to GitHub

Connect repository to Render

Render auto-detects render.yaml

Deploy in minutes

Live URL: https://url-shortener-wtor.onrender.com

🔍 Localhost vs Production
Why does it show localhost when running locally?
When you run the app on your computer with go run cmd/server/main.go, the server binds to localhost:8080. The JavaScript uses window.location.origin to display the URL, which gives http://localhost:8080.

This is correct behavior! Here's why:

Environment	URL shown	Why
Local Development	http://localhost:8080/abc123	The app is running on your machine.
Production (Render)	https://url-shortener-wtor.onrender.com/abc123	The app is running on Render's servers.
The window.location.origin dynamically detects the domain, so the same code works everywhere without changes. When you deploy to Render, it automatically uses the Render domain.

The short link still works in production!
The short code (like on.com) is stored in the database. When someone clicks https://url-shortener-wtor.onrender.com/on.com, the server looks up on.com in its database and redirects to the original URL (e.g., https://amazon.com).

🎯 Future Improvements
3-node cluster for full fault tolerance

Persistent storage with BadgerDB

Custom short codes (e.g., short.ly/google)

Click analytics and statistics

API key authentication

Rate limiting

📄 License
This project is licensed under the MIT License.

🙏 Acknowledgments
Hashicorp Raft – Consensus algorithm

gRPC – RPC framework

Render – Deployment platform

👤 Author
Syed Saaduddin

GitHub: @SyedSaaduddin1101

<p align="center"> <b>Made with ❤️ using Go, Raft, and gRPC</b> </p><p align="center"> ⭐ If you found this useful, star it on GitHub! ⭐ </p> ```
