package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)


type IPResponse struct {
	Status 			string `json:"status"`
	Country 		string `json:"country"`
	RegionName 	string `json:"regionName"`
	City 				string `json:"city"`
	Timezone 		string `json:"timezone"`
	Isp 				string `json:"isp"`
	Org 				string `json:"org"`
	As 					string `json:"as"`
	Query 			string `json:"query"`
	IP 					net.IP `json:"ip"`
}





func main() {
	http.HandleFunc("/", healthCheck)
	http.HandleFunc("/analytics", handleAnalytics)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server Running on Port " + port)
	
	log.Fatal(http.ListenAndServe(":" + port, nil))
}


func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Server Running successfully")
}


func handleAnalytics(w http.ResponseWriter, r *http.Request) {
	// w.Header().Add("Content-Type", "application/json")
	ip := getIP(r)
	log.Println(ip)
	req, err := http.NewRequest("GET", "http://ip-api.com/json/" + ip.String() , nil)
	if err != nil {
		fmt.Fprintf(w, "Failed to create request" + err.Error())
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "Failed to do request" + err.Error())
		return
	}

	decoder := json.NewDecoder(resp.Body)

	var jsonResp IPResponse
	err = decoder.Decode(&jsonResp)
	jsonResp.IP = ip


	storeAnalytics(jsonResp)
	if err != nil {
		fmt.Fprintf(w, "Failed to decode response" + err.Error())
		return
	}
	b, err := io.ReadAll(resp.Body)
if err != nil {
	fmt.Fprint(w, "Failed to read response " + err.Error())
	return
}
	w.Write(b)
}

// Get the IP address of the server's connected user.
func getIP(r *http.Request) net.IP {
	var userIP string
	if len(r.Header.Get("CF-Connecting-IP")) > 1 {
			userIP = r.Header.Get("CF-Connecting-IP")
			return net.ParseIP(userIP)
	} else if len(r.Header.Get("X-Forwarded-For")) > 1 {
			userIP = r.Header.Get("X-Forwarded-For")
			return net.ParseIP(userIP)
	} else if len(r.Header.Get("X-Real-IP")) > 1 {
			userIP = r.Header.Get("X-Real-IP")
			return net.ParseIP(userIP)
	} else {
		ip, port, err := net.SplitHostPort(r.RemoteAddr)
		log.Println(port)
		
    if err != nil {
        fmt.Printf("userip: %q is not IP:port", r.RemoteAddr)
				return nil
    }	
		
    return  net.ParseIP(ip)
	}
}

func storeAnalytics(d IPResponse) {
	// store or send email
}