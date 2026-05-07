package database

import (
	"database/sql"
	"errors"
	"log"
	"strings" // Added for role normalization
	"time"

	_ "github.com/glebarez/go-sqlite"
	"golang.org/x/crypto/bcrypt"
	"proiect_lms/models"
	"proiect_lms/views"
)

var ErrDbNotInitialised = errors.New("Database not initialised!")
var ErrBookUnavailable = errors.New("No available copies for this book!")

type DbManager struct {
	db *sql.DB
}

type SearchContext struct {
	Query        string
	SearchOption string
	IsAdmin      bool
}

var filters = map[string]string{
	"Title":    " title LIKE ?",
	"Author":   " author LIKE ?",
	"Category": " category LIKE ?",
	"ISBN":     " isbn = ?",
	"All":      " 1=1",
}

var loansFilters = map[string]string{
	"Returned":   " return_date NOT NULL",
	"Unreturned": " return_date IS NULL",
	"All":        " 1=1",
}

func InitDB() (*DbManager, error) {
	db, err := sql.Open("sqlite", "./database/lib.db")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected!")
	return &DbManager{db: db}, nil
}

func (dm *DbManager) GetUserAuth(email string, password string) (*views.UserView, error) {
	if dm.db == nil {
		return nil, ErrDbNotInitialised
	}

	var u views.UserView
	var hash string

	row := dm.db.QueryRow(`SELECT user_id, name, email, password_hash, role FROM Users WHERE email = ?`, email)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &hash, &u.Role)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	// Normalize role to ensure proper UI rendering
	if strings.ToLower(u.Role) == "admin" {
		u.Role = "Admin"
	} else {
		u.Role = "User"
	}

	return &u, nil
}

func (dm *DbManager) GetUsers() ([]views.UserView, error) {
	if dm.db == nil {
		return nil, ErrDbNotInitialised
	}

	rows, err := dm.db.Query("SELECT user_id, name, email, role FROM Users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []views.UserView
	for rows.Next() {
		var u views.UserView
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (dm *DbManager) AddUser(u *models.User) (int64, error) {
	if dm.db == nil {
		return 0, ErrDbNotInitialised
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	res, err := dm.db.Exec(
		"INSERT INTO Users (name, email, password_hash, role) VALUES (?, ?, ?, ?)",
		u.Name, u.Email, hash, u.Role)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	log.Printf("User with id %d created!", id)
	return id, nil
}

func (dm *DbManager) SearchBooks(ctx *SearchContext) ([]views.BookView, error) {
	if dm.db == nil {
		return nil, ErrDbNotInitialised
	}

	baseQuery := "SELECT book_id, author, title, isbn, publish_year, category, number_of_copies FROM Books"
	var sqlFilter string
	var args []any

	if ctx.Query == "" {
		sqlFilter = filters["All"]
	} else {
		if ctx.SearchOption == "ISBN" && !ctx.IsAdmin {
			return nil, errors.New("unauthorized search option")
		}
		f, exists := filters[ctx.SearchOption]
		if !exists {
			sqlFilter = filters["All"]
		} else {
			sqlFilter = f
			if ctx.SearchOption == "ISBN" {
				args = append(args, ctx.Query)
			} else {
				args = append(args, "%"+ctx.Query+"%")
			}
		}
	}

	rows, err := dm.db.Query(baseQuery+" WHERE"+sqlFilter, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []views.BookView
	for rows.Next() {
		var b views.BookView
		err = rows.Scan(&b.ID, &b.Author, &b.Title, &b.ISBN, &b.Year, &b.Category, &b.TotalCopies)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (dm *DbManager) AddBook(b *models.Book) (int64, error) {
	if dm.db == nil {
		return 0, ErrDbNotInitialised
	}

	res, err := dm.db.Exec(
		"INSERT INTO Books (author, title, isbn, publish_year, category, number_of_copies) VALUES (?, ?, ?, ?, ?, ?)",
		b.Author, b.Title, b.Isbn, b.Year, b.Category, b.Copies)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	log.Printf("Book with id %d added!", id)
	return id, nil
}

func (dm *DbManager) DeleteBook(book_id int64) error {
	if dm.db == nil {
		return ErrDbNotInitialised
	}

	_, err := dm.db.Exec("DELETE FROM Books WHERE book_id = ?", book_id)
	if err != nil {
		return err
	}

	log.Printf("Book with id %d deleted successfully!", book_id)
	return nil
}

// LoanBook atomically decrements available copies and creates the loan record.
// Returns ErrBookUnavailable if no copies remain.
func (dm *DbManager) LoanBook(user_id int64, book_id int64) (int64, error) {
	if dm.db == nil {
		return 0, ErrDbNotInitialised
	}

	tx, err := dm.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, err := tx.Exec(
		"UPDATE Books SET number_of_copies = number_of_copies - 1 WHERE book_id = ? AND number_of_copies > 0",
		book_id)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		err = ErrBookUnavailable
		return 0, err
	}

	borrowDate := time.Now()
	dueDate := borrowDate.AddDate(0, 0, 14)

	res, err = tx.Exec(
		"INSERT INTO Loans (user_id, book_id, borrow_date, due_date) VALUES (?, ?, ?, ?)",
		user_id, book_id, borrowDate, dueDate)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	log.Printf("Loan %d created: user %d loaned book %d", id, user_id, book_id)
	return id, nil
}

func (dm *DbManager) GetLoans(user_id int64, ctx *SearchContext) ([]views.LoanView, error) {
	if dm.db == nil {
		return nil, ErrDbNotInitialised
	}

	baseQuery := "SELECT loan_id, user_id, book_id, borrow_date, due_date, return_date FROM Loans"

	f, exists := loansFilters[ctx.SearchOption]
	if !exists {
		f = loansFilters["All"]
	}

	sqlFilter := f
	var args []any

	if !ctx.IsAdmin {
		sqlFilter += " AND user_id = ?"
		args = append(args, user_id)
	}

	rows, err := dm.db.Query(baseQuery+" WHERE"+sqlFilter, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []views.LoanView
	for rows.Next() {
		var l views.LoanView
		err = rows.Scan(&l.ID, &l.UserID, &l.BookID, &l.LoanDate, &l.DueDate, &l.ReturnDate)
		if err != nil {
			return nil, err
		}
		loans = append(loans, l)
	}
	return loans, nil
}
