package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

const defaultPort = "8080"

func getMyIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// CreateLabelsServer creates HTTP server which returns labels info in response
func CreateLabelsServer() {
	var host string
	mux := http.NewServeMux()
	mux.HandleFunc("/", viewHandler)
	mux.HandleFunc("/labels", getLabelsForIP)
	//Start listening on port
	port := os.Getenv("LABELS_PORT")
	if len(port) == 0 {
		port = defaultPort
	}
	host = getMyIP()
	log.Printf("Listening on %s:%s...\n", host, port)
	log.Fatal(http.ListenAndServe(host+":"+port, mux))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, You asked for %s!", r.URL.Path[1:])
}

func getLabelsForIP(w http.ResponseWriter, r *http.Request) {
	// Only GET method allowed, return a 405 'Method Not Allowed' response.
	log.Printf("Request came to the getLabelsForIP handler.")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(405), 405)
		return
	}
	log.Printf("getLabelsForIP: HTTP GET request")
	// Retrieve the IP from the request URL query string. If there is
	// no ip key in the query string then Get() will return an empty
	// string. We check for this, returning a 400 Bad Request response
	// if it's missing.
	q := r.URL.Query()
	log.Println("Query:", q)
	ipset := q["ip"]
	log.Println("Requested IPs:", ipset, "length: ", len(ipset))
	if len(ipset) == 0 {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	// Validate that the ip is a valid ip by trying to convert it,
	// returning a 400 Bad Request response if the conversion fails.
	for i, ip := range ipset {
		if err := net.ParseIP(ip); err == nil {
			log.Printf("%d) Invalid IP address:%s\n", i, ip)
			http.Error(w, http.StatusText(400), 400)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if len(ipset) == 1 {
		podinfo, err := GetOneFromDB(ipset[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		log.Printf("Podinfo: %v\n", podinfo)
		json.NewEncoder(w).Encode(podinfo)
		log.Println("Send response")
		//fmt.Fprintf(w, "%s represents %s.%s having labels:%s \n", podinfo.IP, podinfo.Service, podinfo.Namespace, podinfo.Labels)
	} else {
		podsinfo, err := GetMultiFromDB(ipset)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		log.Printf("Pods' Info: %v\n", podsinfo)
		json.NewEncoder(w).Encode(podsinfo)
		var output string
		for _, pi := range podsinfo {
			r := fmt.Sprintf("%s represents %s.%s having labels:%s\t", pi.IP, pi.Service, pi.Namespace, pi.Labels)
			output = output + r
		}
		log.Println("Response for multiple IPs: ", output)
		//fmt.Fprintf(w, "%s\n", output)
	}
}
