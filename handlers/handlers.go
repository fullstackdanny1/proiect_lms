package handlers

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"proiect_lms/database"
	"proiect_lms/models"
	"proiect_lms/templates"
	"strconv"
    "errors"
	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
)

func init() {
	gob.Register(int64(0))
}

type HandlerContext struct {
	Dm    *database.DbManager
	Store *sessions.CookieStore
}

// isAuthenticated now handles HTMX-specific redirects
func (ctx *HandlerContext) isAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	session, _ := ctx.Store.Get(r, "session-id")
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
		return false
	}
	return true
}

func (ctx *HandlerContext) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		emptyNav := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return nil })
		templates.Layout("Login", emptyNav, templates.LoginPage()).Render(r.Context(), w)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := ctx.Dm.GetUserAuth(email, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "Invalid email or password")
			return
		}

		session, _ := ctx.Store.Get(r, "session-id")
		session.Values["user_id"] = user.ID
		session.Values["role"] = user.Role
		session.Values["authenticated"] = true

		// Save session BEFORE redirecting
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Session save failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
	}
}

func (ctx *HandlerContext) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !ctx.isAuthenticated(w, r) {
		return
	}
	session, _ := ctx.Store.Get(r, "session-id")
	role, _ := session.Values["role"].(string)

	nav := templates.Sidebar(role)
	content := templates.Dashboard(role == "Admin")
	templates.Layout("Dashboard", nav, content).Render(r.Context(), w)
}

func (ctx *HandlerContext) SearchBooksHandler(w http.ResponseWriter, r *http.Request) {
	if !ctx.isAuthenticated(w, r) {
		return
	}

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

	session, _ := ctx.Store.Get(r, "session-id")
	role, _ := session.Values["role"].(string)
	isAdmin := role == "Admin"

	searchCtx := database.SearchContext{
		Query:        r.FormValue("search"),
		SearchOption: r.FormValue("filterBy"),
		IsAdmin:      isAdmin,
	}

	books, err := ctx.Dm.SearchBooks(&searchCtx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if len(books) == 0 {
		fmt.Fprintf(w, "<div class='p-4 text-green-600'>No books found!</div>")
		return
	}

	templates.BookListResults(books, isAdmin).Render(r.Context(), w)
}

func (ctx *HandlerContext) AddBookHandler(w http.ResponseWriter, r *http.Request) {
	if !ctx.isAuthenticated(w, r) {
		return
	}

	session, _ := ctx.Store.Get(r, "session-id")
	role, _ := session.Values["role"].(string)

	if role != "Admin" {
		http.Error(w, "Unauthorized: Admin access required", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		nav := templates.Sidebar(role)
		content := templates.AddBookForm()
		templates.Layout("Add New Book", nav, content).Render(r.Context(), w)
		return
	}

	if r.Method == http.MethodPost {
		yearOfPub, _ := strconv.ParseInt(r.FormValue("year"), 10, 64)
		copies, _ := strconv.ParseInt(r.FormValue("copies"), 10, 64)

		b := models.Book{
			Title:    r.FormValue("title"),
			Author:   r.FormValue("author"),
			Category: r.FormValue("category"),
			Isbn:     r.FormValue("isbn"),
			Year:     yearOfPub,
			Copies:   copies,
		}

		if _, err := ctx.Dm.AddBook(&b); err != nil {
			http.Error(w, "Internal error!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
	}
}

func (ctx *HandlerContext) DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	if !ctx.isAuthenticated(w, r) {
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	bookId, _ := strconv.ParseInt(r.FormValue("book_id"), 10, 64)

	if err := ctx.Dm.DeleteBook(bookId); err != nil {
		http.Error(w, "Internal error!", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ctx *HandlerContext) LoanBookHandler(w http.ResponseWriter, r *http.Request) {
	if !ctx.isAuthenticated(w, r) {
		return
	}

	session, _ := ctx.Store.Get(r, "session-id")
	userId, _ := session.Values["user_id"].(int64)
	bookId, _ := strconv.ParseInt(r.FormValue("book_id"), 10, 64)

	if _, err := ctx.Dm.LoanBook(userId, bookId); err != nil {
		if errors.Is(err, database.ErrBookUnavailable) {
			http.Error(w, "No copies available for this book", http.StatusConflict)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ctx *HandlerContext) GetLoansHandler(w http.ResponseWriter, r *http.Request) {
	if !ctx.isAuthenticated(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	session, _ := ctx.Store.Get(r, "session-id")
	role, _ := session.Values["role"].(string)
	userId, _ := session.Values["user_id"].(int64)
	isAdmin := role == "Admin"

	searchContext := database.SearchContext{
		SearchOption: r.FormValue("filterBy"),
		IsAdmin:      isAdmin,
	}
	if isAdmin { userId = 0 }

	loans, err := ctx.Dm.GetLoans(userId, &searchContext)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	isHTMX := r.Header.Get("HX-Request") == "true"
	
	rows := templates.LoanListRows(loans)

	if isHTMX {
		if len(loans) == 0 {
			fmt.Fprintf(w, "<div class='text-orange-500'>No loans found.</div>")
			return
		}
		rows.Render(r.Context(), w)
	} else {
		nav := templates.Sidebar(role)
		pageContent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			fmt.Fprintf(w, "<div class='p-6'> <h1 class='text-2xl mb-4'>Loans List</h1> <table>")
			rows.Render(ctx, w)
			fmt.Fprintf(w, "</table> </div>")
			return nil
		})
		templates.Layout("My Loans", nav, pageContent).Render(r.Context(), w)
	}
}

func (ctx *HandlerContext) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := ctx.Store.Get(r, "session-id")
	session.Options.MaxAge = -1
	w.Header().Set("HX-Redirect", "/login")
	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save error on logout: %v", err)
	}
}
