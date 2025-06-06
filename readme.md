# Notes Mailgun Demo

A simple Go web application demonstrating user authentication, session management, and email notifications using Mailgun.

## Features

- User registration and login
- Session-based authentication (using `gin-contrib/sessions`)
- Flash messages for errors and success
- Sending emails via Mailgun API
- Basic note-taking functionality

## Tech Stack

- Go (Gin web framework)
- Mailgun API for email
- Cookie-based sessions

## Setup

1. Clone the repository:
```bash
   git clone git@github.com:rexfordnyrk/notes_mailgun_demo.git cd notes_mailgun_demo
```
2. Install dependencies:
```bash
   go mod tidy
```
3. Set environment variables:
   - `MAILGUN_API_KEY`: Your Mailgun API key
   - `MAILGUN_DOMAIN`: Your Mailgun domain
   - `SESSION_SECRET`: A secret key for session management

4. Run the application:
```bash
   go run main.go
```

5. Visit `http://localhost:8080` in your browser.

## License

MIT