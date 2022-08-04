package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
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
}

type ActionParams struct {
	Type 				string `json:"type"`
	Source 			string `json:"source"`
	Description string `json:"description"`
}

var knownIps = map[string]string{
  "154.113.68.102": "Abdulkarim's Laptop",
  "102.89.34.128" : "Abdulkarim's Phone",
	"197.210.226.113": "Mariams's Phone",
	"102.89.33.246": "Unknown",
	"102.89.32.245": "Ahmed's Phone",
}




func main() {
	godotenv.Load()
	mux := http.NewServeMux()
	mux.HandleFunc("/", healthCheck)
	mux.HandleFunc("/analytics", handleAnalytics)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server Running on Port " + port)
	
	log.Fatal(http.ListenAndServe(":" + port, cors.Default().Handler(mux)))
}


func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Server Running successfully")
}


func handleAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, "This endpoint expects a post request", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var action ActionParams
	err := decoder.Decode(&action)
	if err != nil {
		http.Error(w, "Failed to decode body" + err.Error(), http.StatusBadRequest)
		return
	}
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

	decoder = json.NewDecoder(resp.Body)

	var jsonResp IPResponse
	err = decoder.Decode(&jsonResp)

	if err != nil {
		fmt.Fprintf(w, "Failed to decode response" + err.Error())
		return
	}

	storeAnalytics(jsonResp, action)
	
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

func storeAnalytics(d IPResponse, action ActionParams) {
	reqTime := time.Now()
	// check if ip is among known ips
	for ip, own := range knownIps {
		if d.Query == ip {
			log.Printf("%q just %q, description:\n %s\n on %q at %q time", own, action.Type, action.Description, action.Source, reqTime)
			return
		}
	}

	b, err := json.MarshalIndent(d, "", "	")
	if err != nil {
		log.Println("Failed to marshal ", err)
	}else {
		log.Println(string(b))
		// send me an email here
		err = sendEmail(b)
		log.Println(err)
	}
}


func sendEmail(msg []byte) error {
  from := os.Getenv("MY_FROM_EMAIL")
  password := os.Getenv("MY_FROM_EMAIL_PASSWORD")

  // Receiver email address.
  to := []string{
    os.Getenv("MY_TO_EMAIL"),
  }

  // smtp server configuration.
  smtpHost := "smtp.gmail.com"
  smtpPort := "587"
  
  // Authentication.
  auth := smtp.PlainAuth("", from, password, smtpHost)
  
  // Sending email.
  err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
  if err != nil {
    return err
  }
  log.Println("Email Sent Successfully!")
	return nil
}