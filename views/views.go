package views

import "time"

type UserView struct {
    ID    int
    Name  string
    Email string
    Role   string 
}

type BookView struct {
    ID          int
    Author      string
    Title       string
    ISBN        string
    Year        int
    Category    string
    TotalCopies int
}

type LoanView struct {
    ID           int
    BookID       string
    UserID       string
    LoanDate     time.Time
    DueDate      time.Time
    ReturnDate *time.Time 
}
