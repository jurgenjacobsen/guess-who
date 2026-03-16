package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
)

//go:embed app/dist
var webDist embed.FS

type Person struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

var peopleFileMutex sync.Mutex

func readPeopleFromFile(filePath string) ([]Person, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Person{}, nil
		}
		return nil, err
	}

	if len(bytes) == 0 {
		return []Person{}, nil
	}

	var people []Person
	if err := json.Unmarshal(bytes, &people); err != nil {
		return nil, err
	}

	if people == nil {
		return []Person{}, nil
	}

	return people, nil
}

func writePeopleToFile(filePath string, people []Person) error {
	bytes, err := json.MarshalIndent(people, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, bytes, 0644)
}

func generatePersonID() string {
	return fmt.Sprintf("person_%d", time.Now().UnixNano())
}

func createStaticHandler() (http.Handler, error) {
	distFS, err := fs.Sub(webDist, "app/dist")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServerFS(distFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if requestPath == "." || requestPath == "" {
			requestPath = "index.html"
		}

		if _, err := fs.Stat(distFS, requestPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		spaRequest := r.Clone(r.Context())
		spaRequest.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, spaRequest)
	}), nil
}

func getPeopleHandler(peopleFilePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		peopleFileMutex.Lock()
		defer peopleFileMutex.Unlock()

		people, err := readPeopleFromFile(peopleFilePath)
		if err != nil {
			http.Error(w, "Failed to read people file", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(people)
	}
}

func addPersonHandler(peopleFilePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var newPerson Person
		if err := json.NewDecoder(r.Body).Decode(&newPerson); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		newPerson.Name = strings.TrimSpace(newPerson.Name)
		newPerson.Photo = strings.TrimSpace(newPerson.Photo)
		if newPerson.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if newPerson.Photo == "" {
			http.Error(w, "photo is required", http.StatusBadRequest)
			return
		}
		newPerson.ID = generatePersonID()

		peopleFileMutex.Lock()
		defer peopleFileMutex.Unlock()

		people, err := readPeopleFromFile(peopleFilePath)
		if err != nil {
			http.Error(w, "Failed to read people file", http.StatusInternalServerError)
			return
		}

		people = append(people, newPerson)
		if err := writePeopleToFile(peopleFilePath, people); err != nil {
			http.Error(w, "Failed to write people file", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(newPerson)
	}
}

func deletePersonHandler(peopleFilePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var requestPerson Person
		if err := json.NewDecoder(r.Body).Decode(&requestPerson); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		requestPerson.ID = strings.TrimSpace(requestPerson.ID)
		if requestPerson.ID == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		peopleFileMutex.Lock()
		defer peopleFileMutex.Unlock()

		people, err := readPeopleFromFile(peopleFilePath)
		if err != nil {
			http.Error(w, "Failed to read people file", http.StatusInternalServerError)
			return
		}

		deleteIndex := -1
		for i, person := range people {
			if person.ID == requestPerson.ID {
				deleteIndex = i
				break
			}
		}

		if deleteIndex == -1 {
			http.Error(w, "person not found", http.StatusNotFound)
			return
		}

		deletedPerson := people[deleteIndex]
		people = append(people[:deleteIndex], people[deleteIndex+1:]...)
		if err := writePeopleToFile(peopleFilePath, people); err != nil {
			http.Error(w, "Failed to write people file", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(deletedPerson)
	}
}

// getLocalIP attempts to find the IPv4 address specifically for an Ethernet adapter.
// If no Ethernet is found, it falls back to the first available non-loopback IPv4.
func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}

	fmt.Println("DEBUG: Searching for Network Adapters starting with 192...")

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 1. Skip if no IP found, is loopback, or is not IPv4
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			ipStr := ip.String()

			// 2. ONLY accept IPs starting with 192
			if strings.HasPrefix(ipStr, "192.") {
				fmt.Printf(">>> MATCHED LOCAL IP: %s (Interface: %s)\n", ipStr, i.Name)
				return ipStr
			}
		}
	}

	fmt.Println("!!! No 192.x.x.x IP found. Falling back to localhost.")
	return "localhost"
}

func main() {
	// 1. Port Selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter port (default 4455): ")
	port, _ := reader.ReadString('\n')

	port = strings.TrimSpace(port)
	if port == "" {
		port = "4455"
	}

	// 2. IP and URL Discovery
	localIP := getLocalIP()
	url := fmt.Sprintf("http://%s:%s", localIP, port)

	// 3. QR Code Generation
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		fmt.Printf("Error generating QR code: %v\n", err)
	} else {
		fmt.Println("\n==========================================")
		fmt.Printf(" SCAN TO CONNECT (Same Wi-Fi/Network)\n")
		fmt.Printf(" URL: %s\n", url)
		fmt.Println("==========================================\n ")
		// false = white on black (standard for dark terminals)
		fmt.Println(q.ToSmallString(false))
	}

	// 4. Static File Server Logic
	peopleFilePath := "./people.json"
	staticHandler, err := createStaticHandler()
	if err != nil {
		fmt.Printf("Error preparing embedded static files: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(peopleFilePath); os.IsNotExist(err) {
		if err := writePeopleToFile(peopleFilePath, []Person{}); err != nil {
			fmt.Printf("Error initializing people file: %v\n", err)
			os.Exit(1)
		}
	}

	http.HandleFunc("/api/people", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPeopleHandler(peopleFilePath)(w, r)
		case http.MethodPost:
			addPersonHandler(peopleFilePath)(w, r)
		case http.MethodDelete:
			deletePersonHandler(peopleFilePath)(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	http.Handle("/", staticHandler)

	// 5. Start Server
	fmt.Printf("\nServer is running at: %s\n", url)
	fmt.Println("Press Ctrl+C to stop.")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("CRITICAL ERROR: %s\n", err)
		os.Exit(1)
	}
}
