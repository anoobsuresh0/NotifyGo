package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req EmailRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	if req.To == "" || req.Subject == "" || req.Body == "" {
		http.Error(w, "Missing email parameters", http.StatusBadRequest)
		return
	}

	// Access email credentials from environment variables
	senderEmail := os.Getenv("EMAIL_SENDER")
	senderPassword := os.Getenv("EMAIL_PASSWORD")

	if senderEmail == "" || senderPassword == "" {
		http.Error(w, "Missing email credentials in environment variables", http.StatusInternalServerError)
		return
	}

	// Construct email message
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", req.To)
	m.SetHeader("Subject", req.Subject)
	m.SetBody("text/plain", req.Body)

	// Configure email dialer
	d := gomail.NewDialer("smtp.gmail.com", 587, senderEmail, senderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Println("Error sending email:", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Email sent successfully")
}

func sendWhatsAppHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var data map[string]string
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, "Error parsing request body", http.StatusBadRequest)
        return
    }

    to, ok := data["to"]
    if !ok {
        http.Error(w, "Missing 'to' parameter in request body", http.StatusBadRequest)
        return
    }

    body, ok := data["body"]
    if !ok {
        http.Error(w, "Missing 'body' parameter in request body", http.StatusBadRequest)
        return
    }

    // Access Twilio credentials from environment variables
    accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
    authToken := os.Getenv("TWILIO_AUTH_TOKEN")
    messagingServiceSid := os.Getenv("TWILIO_MESSAGING_SERVICE_SID")

    if accountSid == "" || authToken == "" || messagingServiceSid == "" {
        http.Error(w, "Missing Twilio credentials in environment variables", http.StatusInternalServerError)
        return
    }

    client := twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: accountSid,
        Password: authToken,
    })

    params := &api.CreateMessageParams{}
    params.SetBody(body)
    params.SetMessagingServiceSid(messagingServiceSid)
    params.SetTo("whatsapp:" + to)

    // Handle optional media URL
    mediaUrl, mediaOk := data["media_url"]
    if mediaOk && mediaUrl != "" {
        params.SetMediaUrl([]string{mediaUrl})
    }

    resp, err := client.Api.CreateMessage(params)
    if err != nil {
        log.Println("Error sending WhatsApp message:", err)
        http.Error(w, "Failed to send WhatsApp message", http.StatusInternalServerError)
        return
    }

    // Log the response from Twilio
    if resp.Body != nil {
        log.Printf("Twilio Response: %s", *resp.Body)
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "WhatsApp message sent successfully")
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/send-email", sendEmailHandler)
	http.HandleFunc("/send-whatsapp", sendWhatsAppHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
