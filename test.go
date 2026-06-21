package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("✅ Received request for:", r.URL.Path)
		w.Write([]byte(`
			<h1>It Works!</h1>
			<p>If you see this, the server is responding correctly.</p>
			<form action="/shorten" method="POST">
				<input type="text" name="url" value="https://google.com">
				<button type="submit">Shorten</button>
			</form>
		`))
	})

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("✅ Received POST to /shorten")
		w.Write([]byte(`{"short_code":"test123"}`))
	})

	fmt.Println("🚀 Test server running at http://localhost:8081")
	fmt.Println("Open http://localhost:8081 in your browser")
	http.ListenAndServe(":8081", nil)
}