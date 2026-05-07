-- Insert Categories
INSERT INTO Categories (category_name) VALUES ('Fiction'), ('Science'), ('History');

-- Insert Users 
INSERT INTO Users (name, email, password_hash, role) VALUES 
('Admin User', 'admin@lib.com', '$2a$10$YourHashedPassHere', 'Admin'),
('Student John', 'john@stud.com', '$2a$10$YourHashedPassHere', 'User');

-- Insert Books
INSERT INTO Books (title, isbn, publish_year, category_id, number_of_copies) VALUES 
('The Great Gatsby', '9780743273565', 1925, 1, 5),
('A Brief History of Time', '9780553380163', 1988, 2, 2);

-- Insert Authors & Link them
INSERT INTO Authors (author_name) VALUES ('F. Scott Fitzgerald'), ('Stephen Hawking');
INSERT INTO BookAuthors (book_id, author_id) VALUES (1, 1), (2, 2);

-- Create a Test Loan 
INSERT INTO Loans (user_id, book_id, due_date) VALUES 
(2, 1, datetime('now', '+14 days'));
