package database

import (
	"fmt"
	"log"
	"proiect_lms/models"
)

// SeedDb populates the database using the manager's built-in methods
func SeedDb(dm *DbManager) {
	log.Println("Seeding database with full dataset using DbManager methods...")

	// 1. Seed 10 Users (3 Admins, 7 Regular)
	// AddUser handles bcrypt hashing automatically
	users := []models.User{
		{Name: "Andrew Admin", Email: "admin@test.com", Password: "password123", Role: "Admin"},
		{Name: "Elena Boss", Email: "elena@test.com", Password: "password123", Role: "Admin"},
		{Name: "Library Manager", Email: "manager@test.com", Password: "password123", Role: "Admin"},
		{Name: "Ion Popescu", Email: "student1@test.com", Password: "password123", Role: "User"},
		{Name: "Maria Ionescu", Email: "student2@test.com", Password: "password123", Role: "User"},
		{Name: "George Enescu", Email: "student3@test.com", Password: "password123", Role: "User"},
		{Name: "Ana Blandiana", Email: "student4@test.com", Password: "password123", Role: "User"},
		{Name: "Mircea Eliade", Email: "student5@test.com", Password: "password123", Role: "User"},
		{Name: "Liviu Rebreanu", Email: "student6@test.com", Password: "password123", Role: "User"},
		{Name: "Mihai Eminescu", Email: "student7@test.com", Password: "password123", Role: "User"},
	}

	var userIDs []int64
	for _, u := range users {
		id, err := dm.AddUser(&u) // Uses AddUser from db.go
		if err != nil {
			log.Printf("Skip seeding user %s (likely already exists): %v", u.Email, err)
			continue
		}
		userIDs = append(userIDs, id)
	}

	// 2. Seed 20 Books
	books := []models.Book{
		{Title: "The Go Programming Language", Author: "Alan Donovan", Category: "Programming", Year: 2015, Copies: 5},
		{Title: "Clean Code", Author: "Robert C. Martin", Category: "Software Engineering", Year: 2008, Copies: 3},
		{Title: "The Pragmatic Programmer", Author: "Andrew Hunt", Category: "Software Engineering", Year: 1999, Copies: 4},
		{Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Category: "Programming", Year: 2017, Copies: 2},
		{Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann", Category: "Architecture", Year: 2017, Copies: 6},
		{Title: "Introduction to Algorithms", Author: "Thomas H. Cormen", Category: "Algorithms", Year: 2009, Copies: 10},
		{Title: "Test Driven Development", Author: "Kent Beck", Category: "Software Engineering", Year: 2002, Copies: 3},
		{Title: "Refactoring", Author: "Martin Fowler", Category: "Software Engineering", Year: 2018, Copies: 5},
		{Title: "Domain-Driven Design", Author: "Eric Evans", Category: "Architecture", Year: 2003, Copies: 2},
		{Title: "Code Complete", Author: "Steve McConnell", Category: "Software Engineering", Year: 2004, Copies: 4},
		{Title: "The Mythical Man-Month", Author: "Fred Brooks", Category: "Management", Year: 1975, Copies: 2},
		{Title: "Structure and Interpretation of Computer Programs", Author: "Harold Abelson", Category: "Programming", Year: 1996, Copies: 3},
		{Title: "Design Patterns", Author: "Erich Gamma", Category: "Architecture", Year: 1994, Copies: 5},
		{Title: "The Clean Architect", Author: "Robert C. Martin", Category: "Architecture", Year: 2017, Copies: 4},
		{Title: "Computer Networking", Author: "James Kurose", Category: "Networking", Year: 2020, Copies: 5},
		{Title: "Modern Operating Systems", Author: "Andrew Tanenbaum", Category: "OS", Year: 2014, Copies: 3},
		{Title: "Artificial Intelligence", Author: "Stuart Russell", Category: "AI", Year: 2020, Copies: 6},
		{Title: "Database System Concepts", Author: "Abraham Silberschatz", Category: "Databases", Year: 2019, Copies: 4},
		{Title: "Compilers: Principles, Techniques, and Tools", Author: "Alfred Aho", Category: "CS", Year: 2006, Copies: 2},
		{Title: "Effective Java", Author: "Joshua Bloch", Category: "Programming", Year: 2018, Copies: 5},
	}

	var bookIDs []int64
	for i, b := range books {
		b.Isbn = fmt.Sprintf("978-0000000%02d", i)
		id, err := dm.AddBook(&b) // Uses AddBook from db.go
		if err != nil {
			log.Printf("Skip seeding book %s: %v", b.Title, err)
			continue
		}
		bookIDs = append(bookIDs, id)
	}

	// 3. Seed 50 Loans
	// Use LoanBook to decrease book inventory and record the transaction
	log.Println("Generating 50 loans using LoanBook method...")
	if len(userIDs) > 3 && len(bookIDs) > 0 {
		regularUserIDs := userIDs[3:] // Use students (IDs generated after admins)
		for i := 0; i < 50; i++ {
			uID := regularUserIDs[i%len(regularUserIDs)]
			bID := bookIDs[i%len(bookIDs)]

			_, err := dm.LoanBook(uID, bID) // Uses LoanBook from db.go
			if err != nil {
				log.Printf("Could not create loan %d: %v", i, err)
			}
		}
	}

	log.Println("Seeding complete.")
}
