
---

## 📸 Screenshots

### 🎨 Web Interface
![Shortify UI](https://via.placeholder.com/800x400/1a1a2e/ff8906?text=Shortify+UI)

*Clean, dark-themed UI with real-time character savings.*

### 📱 Mobile Responsive
![Mobile View](https://via.placeholder.com/400x800/1a1a2e/38bdf8?text=Mobile+Responsive)

*Works perfectly on all devices.*

---

## 🛠️ Tech Stack

| Component | Technology |
| :--- | :--- |
| **Language** | Go 1.21+ |
| **RPC** | gRPC + Protocol Buffers |
| **Consensus** | Hashicorp Raft |
| **Networking** | TCP / HTTP/2 |
| **Container** | Docker + Docker Compose |
| **Deployment** | Render.com |

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
