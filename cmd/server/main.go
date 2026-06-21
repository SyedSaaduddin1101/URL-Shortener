package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"distributed-url-shortener/internal/fsm"
	pb "distributed-url-shortener/proto"

	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
)

// The NEW interactive HTML page
const htmlPage = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Shortify - URL Shortener</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: #0f172a;
            color: #e2e8f0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            background: #1e293b;
            padding: 2.5rem;
            border-radius: 24px;
            width: 600px;
            max-width: 100%;
            box-shadow: 0 20px 60px rgba(0,0,0,0.7);
            border: 1px solid #334155;
        }
        .header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
        }
        .header .logo {
            font-size: 2rem;
        }
        h1 {
            font-size: 2rem;
            font-weight: 700;
            background: linear-gradient(135deg, #38bdf8, #818cf8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .subtitle {
            color: #94a3b8;
            margin-bottom: 2rem;
            font-size: 0.95rem;
        }
        .input-group {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }
        .input-row {
            display: flex;
            gap: 10px;
            width: 100%;
        }
        input[type="url"] {
            flex: 1;
            padding: 14px 18px;
            border: 2px solid #334155;
            border-radius: 12px;
            background: #0f172a;
            color: #f8fafc;
            font-size: 1rem;
            transition: all 0.3s ease;
            outline: none;
            width: 100%;
        }
        input[type="url"]:focus {
            border-color: #38bdf8;
            box-shadow: 0 0 0 4px rgba(56, 189, 248, 0.15);
        }
        button {
            padding: 14px 28px;
            background: linear-gradient(135deg, #38bdf8, #818cf8);
            color: #fff;
            border: none;
            border-radius: 12px;
            font-weight: 600;
            font-size: 1rem;
            cursor: pointer;
            transition: all 0.3s ease;
            white-space: nowrap;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(56, 189, 248, 0.3);
        }
        button:active {
            transform: scale(0.97);
        }
        button:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        
        .loader {
            display: none;
            margin-top: 20px;
            text-align: center;
            color: #94a3b8;
        }
        .loader-spinner {
            display: inline-block;
            width: 24px;
            height: 24px;
            border: 3px solid #334155;
            border-top-color: #38bdf8;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        
        .result {
            margin-top: 24px;
            background: #0f172a;
            border-radius: 16px;
            padding: 20px;
            border-left: 4px solid #38bdf8;
            display: none;
            animation: slideUp 0.4s ease forwards;
        }
        .result.show { display: block; }
        
        @keyframes slideUp {
            from { opacity: 0; transform: translateY(15px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        .result .label {
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #94a3b8;
            margin-bottom: 4px;
            display: block;
        }
        .url-row {
            display: flex;
            align-items: center;
            gap: 12px;
            background: #1e293b;
            padding: 8px 12px 8px 16px;
            border-radius: 10px;
            border: 1px solid #334155;
            margin-top: 4px;
        }
        .url-row a {
            flex: 1;
            color: #38bdf8;
            text-decoration: none;
            font-weight: 500;
            word-break: break-all;
            font-size: 1.05rem;
        }
        .url-row a:hover { text-decoration: underline; }
        .copy-btn {
            background: #334155;
            border: none;
            color: #e2e8f0;
            padding: 6px 16px;
            border-radius: 8px;
            font-size: 0.8rem;
            font-weight: 600;
            cursor: pointer;
            transition: 0.2s;
            white-space: nowrap;
        }
        .copy-btn:hover {
            background: #475569;
            transform: scale(1.02);
        }
        .copy-btn.copied {
            background: #22c55e;
            color: #fff;
        }
        .original-section {
            margin-top: 16px;
            padding-top: 16px;
            border-top: 1px solid #334155;
        }
        .original-section p {
            color: #94a3b8;
            font-size: 0.9rem;
            word-break: break-all;
        }
        .stats {
            margin-top: 12px;
            display: flex;
            gap: 20px;
            font-size: 0.85rem;
        }
        .stats span {
            background: #1e293b;
            padding: 4px 12px;
            border-radius: 20px;
            border: 1px solid #334155;
            color: #cbd5e1;
        }
        .stats .saved {
            color: #4ade80;
            border-color: #4ade80;
        }
        .error {
            margin-top: 16px;
            color: #f87171;
            background: #7f1d1d33;
            padding: 12px;
            border-radius: 10px;
            border-left: 4px solid #f87171;
            display: none;
        }
        .error.show { display: block; }
        .footer {
            margin-top: 24px;
            text-align: center;
            font-size: 0.75rem;
            color: #475569;
        }
        .footer a { color: #64748b; text-decoration: none; }
    </style>
</head>
<body>
<div class="container">
    <div class="header">
        <span class="logo">⚡</span>
        <h1>Shortify</h1>
    </div>
    <p class="subtitle">Instantly shorten your long, ugly URLs.</p>

    <div class="input-group">
        <div class="input-row">
            <input type="url" id="urlInput" placeholder="Paste your long URL here..." value="https://google.com">
            <button id="shortenBtn" onclick="shorten()">Shorten</button>
        </div>
    </div>

    <div class="loader" id="loader">
        <div class="loader-spinner"></div>
        <span style="display:block; margin-top: 8px;">Shortening...</span>
    </div>

    <div class="error" id="error"></div>

    <div class="result" id="result">
        <span class="label">Your Short Link</span>
        <div class="url-row">
            <a href="#" id="shortLink" target="_blank">...</a>
            <button class="copy-btn" id="copyBtn" onclick="copyUrl()">Copy</button>
        </div>
        <div class="original-section">
            <span class="label">Original</span>
            <p id="originalDisplay">...</p>
        </div>
        <div class="stats">
            <span id="charCount"></span>
            <span class="saved" id="savedChars"></span>
        </div>
    </div>
    <div class="footer">⚡ Powered by Raft Consensus &amp; gRPC</div>
</div>

<script>
    let lastShortUrl = '';

    async function shorten() {
        const url = document.getElementById('urlInput').value.trim();
        const resultDiv = document.getElementById('result');
        const errorDiv = document.getElementById('error');
        const loaderDiv = document.getElementById('loader');
        const btn = document.getElementById('shortenBtn');

        if (!url) {
            errorDiv.textContent = 'Please enter a URL.';
            errorDiv.classList.add('show');
            return;
        }
        // Basic URL validation
        if (!url.startsWith('http://') && !url.startsWith('https://')) {
            errorDiv.textContent = 'Enter a valid URL (include http:// or https://)';
            errorDiv.classList.add('show');
            return;
        }

        errorDiv.classList.remove('show');
        resultDiv.classList.remove('show');
        loaderDiv.style.display = 'block';
        btn.disabled = true;

        try {
            const res = await fetch('/shorten', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url: url })
            });
            const data = await res.json();

            if (data.short_code) {
                const shortUrl = window.location.origin + '/' + data.short_code;
                lastShortUrl = shortUrl;

                document.getElementById('shortLink').href = shortUrl;
                document.getElementById('shortLink').textContent = shortUrl;
                document.getElementById('originalDisplay').textContent = url;

                const saved = url.length - shortUrl.length;
                document.getElementById('charCount').textContent = '📏 ' + url.length + ' chars';
                document.getElementById('savedChars').textContent = '✅ Saved ' + saved + ' chars';

                resultDiv.classList.add('show');
                
                // Reset copy button
                const copyBtn = document.getElementById('copyBtn');
                copyBtn.textContent = 'Copy';
                copyBtn.classList.remove('copied');
            } else {
                errorDiv.textContent = 'Error: ' + (data.error || 'Something went wrong');
                errorDiv.classList.add('show');
            }
        } catch (err) {
            errorDiv.textContent = 'Network error. Is the server running?';
            errorDiv.classList.add('show');
        } finally {
            loaderDiv.style.display = 'none';
            btn.disabled = false;
        }
    }

    function copyUrl() {
        const copyBtn = document.getElementById('copyBtn');
        if (!lastShortUrl) return;
        navigator.clipboard.writeText(lastShortUrl).then(() => {
            copyBtn.textContent = 'Copied!';
            copyBtn.classList.add('copied');
            setTimeout(() => {
                copyBtn.textContent = 'Copy';
                copyBtn.classList.remove('copied');
            }, 2000);
        }).catch(() => {
            // Fallback for older browsers
            const input = document.createElement('input');
            input.value = lastShortUrl;
            document.body.appendChild(input);
            input.select();
            document.execCommand('copy');
            document.body.removeChild(input);
            copyBtn.textContent = 'Copied!';
            setTimeout(() => copyBtn.textContent = 'Copy', 2000);
        });
    }

    // Allow Enter key
    document.getElementById('urlInput').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') shorten();
    });
</script>
</body>
</html>`

func main() {
	// Create data folder for Raft
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data folder: %v", err))
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node1"
	}
	raftPort := os.Getenv("RAFT_PORT")
	if raftPort == "" {
		raftPort = "8000"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	store := fsm.New()

	// Setup Raft
	raftAddr := "127.0.0.1:" + raftPort
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.bolt"))
	if err != nil {
		panic(fmt.Sprintf("Failed to create log store: %v", err))
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.bolt"))
	if err != nil {
		panic(fmt.Sprintf("Failed to create stable store: %v", err))
	}
	snapshots, err := raft.NewFileSnapshotStore(dataDir, 3, os.Stderr)
	if err != nil {
		panic(fmt.Sprintf("Failed to create snapshot store: %v", err))
	}
	transport, err := raft.NewTCPTransport(raftAddr, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		panic(fmt.Sprintf("Failed to create transport: %v", err))
	}

	r, err := raft.NewRaft(config, store, logStore, stableStore, snapshots, transport)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Raft: %v", err))
	}

	// Single node cluster for local testing
	if nodeID == "node1" {
		r.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: "node1", Address: transport.LocalAddr()},
			},
		})
	}

	// Create HTTP mux
	mux := http.NewServeMux()

	// Root handler - serves HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// If it's not exactly "/", treat as short code redirect
		if req.URL.Path != "/" {
			code := req.URL.Path[1:]
			longURL, found := store.Resolve(code)
			if !found {
				http.NotFound(w, req)
				return
			}
			http.Redirect(w, req, longURL, http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlPage))
	})

	// Shorten handler
	mux.HandleFunc("/shorten", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		var body struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}

		if r.State() != raft.Leader {
			http.Error(w, "Not the leader", 503)
			return
		}

		// Generate short code (last 6 characters)
		code := body.URL
		if len(code) > 6 {
			code = code[len(code)-6:]
		} else {
			code = "abc123"
		}

		// Send command to Raft
		cmd := fsm.Command{Op: "shorten", Code: code, URL: body.URL}
		data, _ := json.Marshal(cmd)
		future := r.Apply(data, 5*time.Second)
		if err := future.Error(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"short_code": code})
	})

	// Start HTTP server
	go func() {
		fmt.Println("🌐 UI available at http://localhost:8080")
		fmt.Println("📱 Open this in your browser")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			panic(fmt.Sprintf("HTTP server failed: %v", err))
		}
	}()

	// GRPC SERVER
	grpcSrv := grpc.NewServer()
	pb.RegisterURLShortenerServer(grpcSrv, &GRPCServer{Raft: r, FSM: store})
	lis, err := net.Listen("tcp", "0.0.0.0:"+grpcPort)
	if err != nil {
		panic(fmt.Sprintf("Failed to listen on gRPC port: %v", err))
	}
	fmt.Printf("🚀 gRPC Node %s ready on :%s\n", nodeID, grpcPort)
	fmt.Println("==========================================")
	fmt.Println("✅ Server is running!")
	fmt.Println("==========================================")
	if err := grpcSrv.Serve(lis); err != nil {
		panic(fmt.Sprintf("gRPC server failed: %v", err))
	}
}

// GRPC SERVER
type GRPCServer struct {
	pb.UnimplementedURLShortenerServer
	Raft *raft.Raft
	FSM  *fsm.URLStore
}

func (s *GRPCServer) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	if s.Raft.State() != raft.Leader {
		return nil, fmt.Errorf("not leader")
	}
	code := req.LongUrl
	if len(code) > 6 {
		code = code[len(code)-6:]
	} else {
		code = "abc123"
	}
	cmd := fsm.Command{Op: "shorten", Code: code, URL: req.LongUrl}
	data, _ := json.Marshal(cmd)
	future := s.Raft.Apply(data, 5*time.Second)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return &pb.ShortenResponse{ShortCode: code}, nil
}

func (s *GRPCServer) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveResponse, error) {
	if s.Raft.State() != raft.Leader {
		return nil, fmt.Errorf("not leader")
	}
	url, found := s.FSM.Resolve(req.ShortCode)
	return &pb.ResolveResponse{LongUrl: url, Found: found}, nil
}