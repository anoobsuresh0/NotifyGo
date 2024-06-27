package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
}

type WhatsAppRequest struct {
	To string `json:"to"`
}

type CombinedRequest struct {
	Email    EmailRequest    `json:"email"`
	WhatsApp WhatsAppRequest `json:"whatsapp"`
	Body     string          `json:"body"`
	MediaUrl string          `json:"media_url,omitempty"`
}

func downloadFile(URL, fileName string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return fileName, nil
}

func sendEmail(req EmailRequest, body string, mediaUrl string) error {
	// Access email credentials from environment variables
	senderEmail := os.Getenv("EMAIL_SENDER")
	senderPassword := os.Getenv("EMAIL_PASSWORD")

	if senderEmail == "" || senderPassword == "" {
		return fmt.Errorf("missing email credentials in environment variables")
	}

	// Construct email message
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", req.To)
	m.SetHeader("Subject", req.Subject)
	m.SetBody("text/plain", body)

	// Attach media if provided
	if mediaUrl != "" {
		u, err := url.Parse(mediaUrl)
		if err != nil {
			return fmt.Errorf("invalid media URL: %v", err)
		}
		fileName := filepath.Base(u.Path)
		filePath, err := downloadFile(mediaUrl, fileName)
		if err != nil {
			return fmt.Errorf("error downloading media: %v", err)
		}
		m.Attach(filePath)
		defer os.Remove(filePath) // Clean up the downloaded file
	}

	// Configure email dialer
	d := gomail.NewDialer("smtp.gmail.com", 587, senderEmail, senderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

func sendWhatsApp(req WhatsAppRequest, body string, mediaUrl string) error {
	// Access Twilio credentials from environment variables
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	messagingServiceSid := os.Getenv("TWILIO_MESSAGING_SERVICE_SID")

	if accountSid == "" || authToken == "" || messagingServiceSid == "" {
		return fmt.Errorf("missing Twilio credentials in environment variables")
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &api.CreateMessageParams{}
	params.SetBody(body)
	params.SetMessagingServiceSid(messagingServiceSid)
	params.SetTo("whatsapp:" + req.To)

	if mediaUrl != "" {
		params.SetMediaUrl([]string{mediaUrl})
	}

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("error sending WhatsApp message: %v", err)
	}

	// Log the response from Twilio
	if resp.Body != nil {
		log.Printf("Twilio Response: %s", *resp.Body)
	}

	return nil
}

func sendCombinedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req CombinedRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	// Send email
	err = sendEmail(req.Email, req.Body, req.MediaUrl)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Send WhatsApp message
	err = sendWhatsApp(req.WhatsApp, req.Body, req.MediaUrl)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to send WhatsApp message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Email and WhatsApp message sent successfully")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/send-message", sendCombinedHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
