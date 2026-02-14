package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var templates *template.Template

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Time     string `json:"time"`
}

func main() {
	// Load .env file if exists
	loadEnvFile(".env")

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "quickcart")
	dbPass := getEnv("DB_PASSWORD", "quickcart123")
	dbName := getEnv("DB_NAME", "quickcart")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
	} else {
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Initialize database
		if err := initDB(); err != nil {
			log.Printf("Warning: Could not initialize database: %v", err)
		}
	}

	// Parse templates
	templates = template.Must(template.New("").Parse(indexHTML))

	// Routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/products", handleProducts)
	http.HandleFunc("/api/error", handleError)
	http.HandleFunc("/api/slow", handleSlow)

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Printf("Open http://localhost:%s in your browser", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // .env file not found, skip silently
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	log.Println("Loaded environment from .env file")
}

func initDB() error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	// Create products table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// Insert sample data if table is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		sampleProducts := []Product{
			{Name: "Laptop", Price: 999.99, Stock: 50},
			{Name: "Smartphone", Price: 699.99, Stock: 100},
			{Name: "Headphones", Price: 149.99, Stock: 200},
			{Name: "Keyboard", Price: 79.99, Stock: 150},
			{Name: "Mouse", Price: 49.99, Stock: 300},
		}

		for _, p := range sampleProducts {
			_, err := db.Exec("INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)",
				p.Name, p.Price, p.Stock)
			if err != nil {
				return err
			}
		}
		log.Println("Sample products inserted")
	}

	return nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	products := []Product{}
	if db != nil {
		rows, err := db.Query("SELECT id, name, price, stock FROM products ORDER BY id")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var p Product
				if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err == nil {
					products = append(products, p)
				}
			}
		}
	}

	templates.ExecuteTemplate(w, "index", products)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "healthy",
		Time:   time.Now().Format(time.RFC3339),
	}

	if db != nil {
		if err := db.Ping(); err != nil {
			response.Database = "disconnected"
		} else {
			response.Database = "connected"
		}
	} else {
		response.Database = "not configured"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "Database not connected", http.StatusServiceUnavailable)
		return
	}

	rows, err := db.Query("SELECT id, name, price, stock FROM products ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// handleError returns a 500 error for testing monitoring/alerting
// Usage: GET /api/error
func handleError(w http.ResponseWriter, r *http.Request) {
	log.Println("Error endpoint called - returning 500")
	http.Error(w, `{"error": "Internal Server Error", "message": "This is a test error for monitoring"}`, http.StatusInternalServerError)
}

// handleSlow simulates a slow response for testing latency monitoring
// Usage: GET /api/slow?delay=2000 (delay in milliseconds, default 3000)
func handleSlow(w http.ResponseWriter, r *http.Request) {
	delayMs := 3000 // default 3 seconds
	if d := r.URL.Query().Get("delay"); d != "" {
		if parsed, err := fmt.Sscanf(d, "%d", &delayMs); err == nil && parsed == 1 {
			if delayMs > 30000 {
				delayMs = 30000 // max 30 seconds
			}
		}
	}

	log.Printf("Slow endpoint called - sleeping for %dms", delayMs)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Slow response completed",
		"delay_ms": delayMs,
	})
}

const indexHTML = `{{define "index"}}<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>QuickCart - Sample Store</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            min-height: 100vh;
        }
        header {
            background: #2563eb;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        header h1 { font-size: 1.5rem; }
        main { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        .status {
            background: white;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 2rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .status-item {
            display: inline-block;
            margin-right: 2rem;
            font-size: 0.9rem;
        }
        .status-dot {
            display: inline-block;
            width: 8px;
            height: 8px;
            border-radius: 50%;
            margin-right: 0.5rem;
        }
        .status-dot.green { background: #22c55e; }
        .status-dot.yellow { background: #eab308; }
        h2 { margin-bottom: 1rem; color: #333; }
        .products {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 1.5rem;
        }
        .product {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .product:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        .product h3 { color: #333; margin-bottom: 0.5rem; }
        .product .price {
            font-size: 1.25rem;
            font-weight: 600;
            color: #2563eb;
            margin-bottom: 0.5rem;
        }
        .product .stock {
            font-size: 0.85rem;
            color: #666;
        }
        .no-products {
            text-align: center;
            padding: 3rem;
            color: #666;
        }
        footer {
            text-align: center;
            padding: 2rem;
            color: #666;
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <header>
        <h1>QuickCart</h1>
    </header>
    <main>
        <div class="status">
            <span class="status-item">
                <span class="status-dot green"></span>
                Server: Running
            </span>
            <span class="status-item">
                <span class="status-dot {{if .}}green{{else}}yellow{{end}}"></span>
                Database: {{if .}}Connected{{else}}Not Connected{{end}}
            </span>
        </div>

        <h2>Products</h2>
        {{if .}}
        <div class="products">
            {{range .}}
            <div class="product">
                <h3>{{.Name}}</h3>
                <div class="price">${{printf "%.2f" .Price}}</div>
                <div class="stock">Stock: {{.Stock}} units</div>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="no-products">
            <p>No products available. Connect to database to see products.</p>
        </div>
        {{end}}
    </main>
    <footer>
        <p>QuickCart Sample Application @TANDIGITAL</p>
    </footer>
</body>
</html>{{end}}`
