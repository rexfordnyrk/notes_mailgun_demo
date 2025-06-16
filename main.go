package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"html/template"
	"log"
	"strings"
)

type User struct {
	gorm.Model
	Email            string `gorm:"unique"`
	PasswordHash     string
	Verified         bool
	VerificationCode string
	ResetToken       string
}

type Note struct {
	gorm.Model
	UserID  uint
	Title   string
	Content string
}

var db *gorm.DB

func preview6(s string) string {
	words := strings.Fields(s)
	if len(words) > 6 {
		return strings.Join(words[:6], " ") + "..."
	}
	return s
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("notes.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Note{})

	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("user_session", store))
	r.SetFuncMap(template.FuncMap{
		"safe":     func(s string) template.HTML { return template.HTML(s) },
		"preview6": preview6,
		"add":      func(a, b int) int { return a + b },
	})
	r.LoadHTMLGlob("pages/*")

	r.GET("/", showHome)
	r.GET("/register", showRegister)
	r.POST("/register", register)
	r.GET("/verify", showVerify)
	r.POST("/verify", verify)
	r.GET("/login", showLogin)
	r.POST("/login", login)
	r.GET("/forgot", showForgot)
	r.POST("/forgot", forgot)
	r.GET("/reset", showReset)
	r.POST("/reset", reset)
	r.GET("/logout", logout)
	r.GET("/notes", authMiddleware, showNotes)
	r.POST("/notes", authMiddleware, createNote)
	r.POST("/notes/:id/edit", authMiddleware, editNote)
	r.POST("/notes/:id/delete", authMiddleware, deleteNote)
	r.POST("/notes/share", authMiddleware, shareNoteHandler)
	r.GET("/notes/public/:id", showPublicNote)

	// Add bulk notification routes (admin only)
	r.GET("/admin/bulk-notify", adminMiddleware, showBulkNotify)
	r.POST("/admin/bulk-notify", adminMiddleware, sendBulkNotification)

	r.Run(":8080")
}
