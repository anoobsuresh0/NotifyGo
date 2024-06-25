package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

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

func createMultipartRequest(to, body string, mediaPath string) (*http.Request, error) {
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	err := writer.WriteField("To", "whatsapp:"+to)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("From", "whatsapp:+14155238886") // Replace with your Twilio WhatsApp number
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("Body", body)
	if err != nil {
		return nil, err
	}

	if mediaPath != "" {
		file, err := os.Open(mediaPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("MediaUrl", filepath.Base(file.Name()))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Access Twilio credentials from environment variables
	
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	if accountSid == "" || authToken == "" {
		return nil, fmt.Errorf("Missing Twilio credentials in environment variables")
	}

	// Twilio API endpoint URL
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/ACd39d8822751796be730b892865faff0f/Messages.json"

	req, err := http.NewRequest("POST", urlStr, bodyBuffer)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(accountSid, authToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
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

	mediaPath, ok := data["media"] // Optional media path

	req, err := createMultipartRequest(to, body, mediaPath)
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the WhatsApp message
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending WhatsApp message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	fmt.Fprintf(w, "WhatsApp message sent successfully! Response Status: %s", resp.Status)
}

func main() {
	http.HandleFunc("/send-email", sendEmailHandler)
	http.HandleFunc("/send-whatsapp", sendWhatsAppHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
