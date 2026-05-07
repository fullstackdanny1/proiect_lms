PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS Users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT CHECK(role IN ('User', 'Admin')) NOT NULL
);

CREATE TABLE IF NOT EXISTS Books (
    book_id INTEGER PRIMARY KEY AUTOINCREMENT,
    author INTEGER NOT NULL,
    title TEXT NOT NULL,
    isbn TEXT UNIQUE NOT NULL,
    publish_year INTEGER,
    category INTEGER,
    number_of_copies INTEGER DEFAULT 1
);


CREATE TABLE IF NOT EXISTS Loans (
    loan_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    book_id INTEGER,
    borrow_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    due_date DATETIME NOT NULL,
    return_date DATETIME, 
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    FOREIGN KEY (book_id) REFERENCES Books(book_id)
);

CREATE TABLE IF NOT EXISTS Fines (
    fine_id INTEGER PRIMARY KEY AUTOINCREMENT,
    loan_id INTEGER,
    suma REAL NOT NULL,
    status TEXT CHECK(Status IN ('Plătit', 'Neplătit')) DEFAULT 'Neplătit',
    FOREIGN KEY (loan_id) REFERENCES Loans(loan_id)
);
