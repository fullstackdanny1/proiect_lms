package main

import (
	"log"
	"net/http"
	"proiect_lms/database"
	"proiect_lms/handlers"

	"github.com/gorilla/sessions"
)

func main() {
	dm, err := database.InitDB()
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	// Use a fixed key for development consistency
	store := sessions.NewCookieStore([]byte("secret-key-12345"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,  // Security: JS cannot read this
		SameSite: http.SameSiteLaxMode,
	}

	ctx := handlers.HandlerContext{
		Dm:    dm,
		Store: store,
	}

	mux := http.NewServeMux()
	
	// Static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mux.HandleFunc("/login", ctx.LoginHandler)
	mux.HandleFunc("/dashboard", ctx.DashboardHandler)
	mux.HandleFunc("/logout", ctx.LogoutHandler)
	mux.HandleFunc("/search-books", ctx.SearchBooksHandler)
	mux.HandleFunc("/add-book", ctx.AddBookHandler)
	mux.HandleFunc("/delete-book", ctx.DeleteBookHandler)
	mux.HandleFunc("/loan-book", ctx.LoanBookHandler)
	mux.HandleFunc("/get-loans", ctx.GetLoansHandler)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
