# NotifyGo
A web app made using golang to send message in whatsapp and mail


Web Reques using postman


## For Email

```bash
POST : http://localhost:8080/send-email
```

## Email body JSON raw
```bash
{
  "to": "anoob.suresh@perleybrook.com",
  "subject": "Trial Subject",
  "body": "Hope this email finds you well , This is a test mail"
}
```


## For Whatsapp
```bash
POST : http://localhost:8080/send-whatsapp
```

## Postman body JSON raw
```bash
{
  "to": "whatsapp:+918075802343", 
  "body": "Your WhatsApp message"
  "media": "path" (optional)
}

```
## Deployment

To deploy this project run

```bash
  go run main.go
```
or 
```bash
  python3 main.py
```


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

`SENDER_EMAIL`

`PASSWORD`

`TWILIO_AUTH_TOKEN `

`TWILIO_ACCOUNT_SID`

The `PASSWORD` used is App password in gmail got from link below 

https://myaccount.google.com/apppasswords
## Steps to run 

- Add the .env variables

- Create the companies.xlxs table
- Add the resume.pdf
- Run locally using python main.py
## Run Locally

Clone the project

```bash
  git clone https://github.com/anoobsuresh0/NotifyGo.git
```

Go to the project directory

```bash
  cd NotifyGo
```

Start the server

```bash
  go run main.go
```

