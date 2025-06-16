package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	sessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func showHome(c *gin.Context) {
	userID, loggedIn := getSessionUserID(c)
	var email string
	if loggedIn {
		var user User
		if err := db.First(&user, userID).Error; err == nil {
			email = user.Email
		}
	}
	// Clear success message after displaying it
	session := sessions.Default(c)
	success := session.Get("success")
	errors := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	fmt.Println("Success:", success)
	fmt.Println("Errors:", errors)
	// Render the home page with user information
	c.HTML(http.StatusOK, "home.html", gin.H{
		"loggedIn": loggedIn,
		"email":    email,
		"success":  success,
		"error":    errors,
	})
}

// showRegister renders the registration page
func showRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

// register handles user registration by creating a new user account
// and generating a verification code
func register(c *gin.Context) {
	email := c.PostForm("email")
	pass := c.PostForm("password")
	confirm := c.PostForm("confirm")
	if pass != confirm {
		c.String(http.StatusBadRequest, "Passwords do not match")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	user := User{Email: email, PasswordHash: string(hash), VerificationCode: code}
	if err := db.Create(&user).Error; err != nil {
		c.String(http.StatusBadRequest, "Email already registered")
		return
	}
	// In your existing registration handler, replace the TODO line with:
	err := sendVerificationEmail(email, code)
	if err != nil {
		log.Printf("Failed to send verification email to %s: %v", email, err)
		// You might want to handle this error appropriately
		// For now, we'll continue with registration but log the error
	}

	c.Redirect(http.StatusSeeOther, "/verify?email="+email)
}

// showVerify renders the verification page with the user's email
func showVerify(c *gin.Context) {
	c.HTML(http.StatusOK, "verify.html", gin.H{"email": c.Query("email")})
}

// verify validates the user's verification code and marks the account as verified
func verify(c *gin.Context) {
	email := c.PostForm("email")
	code := c.PostForm("code")
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "verify.html", gin.H{"error": "User not found", "email": email})
		return
	}
	if user.VerificationCode == code {
		user.Verified = true
		db.Save(&user)
		session := sessions.Default(c)
		session.Set("success", "Account verified successfully! Please login.")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/login")
	} else {
		c.HTML(http.StatusBadRequest, "verify.html", gin.H{"error": "Invalid code", "email": email})
	}
}

// showLogin renders the login page
func showLogin(c *gin.Context) {
	session := sessions.Default(c)
	success := session.Get("success")
	error := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()
	c.HTML(http.StatusOK, "login.html", gin.H{"success": success, "error": error, "email": c.Query("email")})
}

// login authenticates the user and creates a new session
func login(c *gin.Context) {
	email := c.PostForm("email")
	pass := c.PostForm("password")
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Invalid credentials", "email": email})
		return
	}
	if !user.Verified {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Account not verified", "email": email})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pass)); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Invalid credentials", "email": email})
		return
	}
	setSession(c, user.ID)
	c.Redirect(http.StatusSeeOther, "/notes")
}

// logout clears the user's session and redirects to the login page
func logout(c *gin.Context) {
	clearSession(c)
	c.Redirect(http.StatusSeeOther, "/login")
}

// showForgot renders the forgot password page
func showForgot(c *gin.Context) {
	c.HTML(http.StatusOK, "forgot.html", nil)
}

// forgot handles password reset requests by generating and storing a reset token
func forgot(c *gin.Context) {
	email := c.PostForm("email")
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.HTML(http.StatusExpectationFailed, "forgot.html", gin.H{"error": "This account does not exist.", "email": email})
		return
	}
	token := fmt.Sprintf("%x", rand.Int63())
	user.ResetToken = token
	db.Save(&user)
	// TODO: Send reset link via email: /reset?token=...
	// In your existing forgot password handler, replace the TODO line with:
	err := sendPasswordResetEmail(email, token)
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", email, err)
		// Handle the error appropriately - you might want to show a generic success message
		// to prevent email enumeration attacks, even if the email fails to send
	}
	session := sessions.Default(c)
	session.Set("success", "A reset link has be sent to your email")
	session.Save()
	c.Redirect(http.StatusSeeOther, "/login")
}

// showReset renders the password reset page with the reset token
func showReset(c *gin.Context) {
	token := c.Query("token")
	var user User
	if err := db.Where("reset_token = ?", token).First(&user).Error; err != nil || token == "" {
		session := sessions.Default(c)
		session.Set("error", "Invalid or expired token")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	c.HTML(http.StatusOK, "reset.html", gin.H{"token": token})
}

// reset handles password reset by validating the token and updating the password
func reset(c *gin.Context) {
	token := c.PostForm("token")
	pass := c.PostForm("password")
	confirm := c.PostForm("confirm")
	if pass != confirm {
		c.HTML(http.StatusBadRequest, "reset.html", gin.H{"error": "Passwords do not match", "token": token})
		return
	}
	var user User
	if err := db.Where("reset_token = ?", token).First(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "reset.html", gin.H{"error": "Invalid or expired token", "token": token})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.ResetToken = ""
	db.Save(&user)
	c.HTML(http.StatusOK, "login.html", gin.H{"success": "Password reset successful!"})
}

// showNotes displays all notes belonging to the authenticated user
func showNotes(c *gin.Context) {
	uid, _ := c.Get("userID")
	var notes []Note
	db.Where("user_id = ?", uid).Find(&notes)
	c.HTML(http.StatusOK, "notes.html", gin.H{"notes": notes, "userID": uid})
}

// createNote adds a new note for the authenticated user
func createNote(c *gin.Context) {
	uid, _ := c.Get("userID")
	title := c.PostForm("title")
	content := c.PostForm("content")
	if len(title) > 255 {
		title = title[:255]
	}
	if len(content) > 10000 {
		content = content[:10000]
	}
	note := Note{UserID: uid.(uint), Title: title, Content: content}
	db.Create(&note)
	c.Redirect(http.StatusSeeOther, "/notes")
}

// editNote updates an existing note's title and content
func editNote(c *gin.Context) {
	uid, _ := c.Get("userID")
	id := c.Param("id")
	var note Note
	if err := db.First(&note, id).Error; err != nil || note.UserID != uid.(uint) {
		c.String(http.StatusForbidden, "Not allowed")
		return
	}
	title := c.PostForm("title")
	content := c.PostForm("content")
	if len(title) > 255 {
		title = title[:255]
	}
	if len(content) > 10000 {
		content = content[:10000]
	}
	note.Title = title
	note.Content = content
	db.Save(&note)
	c.Redirect(http.StatusSeeOther, "/notes")
}

// deleteNote removes a note
func deleteNote(c *gin.Context) {
	uid, _ := c.Get("userID")
	id := c.Param("id")
	var note Note
	if err := db.First(&note, id).Error; err != nil || note.UserID != uid.(uint) {
		c.String(http.StatusForbidden, "Not allowed")
		return
	}
	db.Delete(&note)
	c.Redirect(http.StatusSeeOther, "/notes")
}

// shareNoteHandler handles sharing a note via email with a public link
func shareNoteHandler(c *gin.Context) {
	uid, _ := c.Get("userID")
	noteID := c.PostForm("note_id")
	emails := c.PostForm("emails")
	var note Note
	if err := db.First(&note, noteID).Error; err != nil || note.UserID != uid.(uint) {
		c.String(http.StatusForbidden, "Not allowed")
		return
	}
	emailList := strings.Split(emails, ",")
	publicLink := c.Request.Host + "/notes/public/" + noteID

	// TODO: Implement actual email sending here.
	// For each email, send the publicLink.
	for i := range emailList {
		emailList[i] = strings.TrimSpace(emailList[i])
		if emailList[i] == "" {
			continue
		}
		// Example: sendEmail(emailList[i], "Shared Note", "View the note: "+publicLink)
		fmt.Println("Sending email to", emailList[i], "with link:", publicLink)
	}

	// Optionally, show a success message
	c.Redirect(http.StatusSeeOther, "/notes")
}

// showPublicNote renders a public view of a note
func showPublicNote(c *gin.Context) {
	id := c.Param("id")
	var note Note
	if err := db.First(&note, id).Error; err != nil {
		c.String(http.StatusNotFound, "Note not found")
		return
	}
	c.HTML(http.StatusOK, "public_note.html", gin.H{
		"title":   note.Title,
		"content": note.Content,
		"created": note.CreatedAt,
		"updated": note.UpdatedAt,
	})
}

// adminMiddleware ensures the user is logged in and is an admin (user ID 1)
func adminMiddleware(c *gin.Context) {
	userID, loggedIn := getSessionUserID(c)
	if !loggedIn || userID != 1 {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	c.Next()
}

// showBulkNotify renders the bulk notification page
func showBulkNotify(c *gin.Context) {
	// Get list of available templates
	templates, err := getAvailableTemplates()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "bulk_notify.html", gin.H{
			"error": "Failed to load email templates",
		})
		return
	}

	c.HTML(http.StatusOK, "bulk_notify.html", gin.H{
		"templates": templates,
	})
}

// sendBulkNotification handles the bulk notification form submission
func sendBulkNotification(c *gin.Context) {
	subject := c.PostForm("subject")
	templateName := c.PostForm("template")
	recipientType := c.PostForm("recipient_type")

	var recipients []string
	var err error

	if recipientType == "all" {
		// Get all user emails from database
		var users []User
		if err := db.Find(&users).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "bulk_notify.html", gin.H{
				"error": "Failed to fetch user emails",
			})
			return
		}
		for _, user := range users {
			recipients = append(recipients, user.Email)
		}
	} else {
		// Get emails from form field
		emails := c.PostForm("specific_emails")
		recipients = strings.Split(emails, ",")
		for i, email := range recipients {
			recipients[i] = strings.TrimSpace(email)
		}
	}

	if len(recipients) == 0 {
		c.HTML(http.StatusBadRequest, "bulk_notify.html", gin.H{
			"error": "No recipients specified",
		})
		return
	}

	// TODO: Send bulk email
	err = sendBulkEmail(recipients, subject, templateName)
	if err != nil {
		// Log the error and show a user-friendly message
		log.Printf("Failed to send bulk email: %v", err)

		c.HTML(http.StatusInternalServerError, "bulk_notify.html", gin.H{
			"error": fmt.Sprintf("Failed to send notifications: %v", err),
		})
		return
	}

	c.HTML(http.StatusOK, "bulk_notify.html", gin.H{
		"success": fmt.Sprintf("Successfully sent notifications to %d recipients", len(recipients)),
	})
}

// getAvailableTemplates returns a list of available email templates
func getAvailableTemplates() ([]string, error) {
	files, err := os.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	var templates []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			templates = append(templates, file.Name())
		}
	}
	return templates, nil
}
