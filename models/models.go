package models

import "time"

type User struct {
   ID       int64  // Added this
   Name     string
   Email    string
   Password string
   Role     string
}

type Book struct {
   ID       int64  // Added this
   Title    string
   Author   string
   Isbn     string
   Year     int64
   Category string
   Copies   int64
}
  
type Loan struct {
   ID         int64 // Added this for easier tracking
   UserId     int64
   BookId     int64
   BorrowDate time.Time
   DueDate    time.Time
   ReturnDate *time.Time
}
