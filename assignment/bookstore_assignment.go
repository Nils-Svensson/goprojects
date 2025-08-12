/************************
Coding Assignment:
*************************/
//A unit test is failing, fix the issue in the code and make all tests pass
//Add a new endpoint to list one book by title
//Include a price field to the books struct

/***********************
Installation
***********************/
//Install Go: https://go.dev/doc/install
//Getting started: https://go.dev/doc/tutorial/getting-started
//to run the code: go run .
//to run the tests: go test

/***********************
Good Go syntax reference
************************/
//https://gobyexample.com/

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
)

func main() {
	store := NewStore()
	store.fillupStore()

	r := mux.NewRouter()

	r.HandleFunc("/books/{title}", store.GetBookByTitle).Methods("GET")
	r.HandleFunc("/books", store.ListBooksAsJson).Methods("GET")

	log.Println("Server listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))

}

type Book struct {
	Author string
	Title  string
	Score  int
	Price  float64
}

type Store struct {
	Books map[string]Book
}

func NewStore() Store {
	books := map[string]Book{}
	return Store{Books: books}
}

func (s *Store) fillupStore() {
	book1 := Book{
		Author: "Stephen King",
		Title:  "the long walk",
		Score:  2,
		Price:  9.99,
	}

	book2 := Book{
		Author: "Andy Weir",
		Title:  "the martian",
		Score:  5,
		Price:  20.00,
	}

	book3 := Book{
		Author: "Isaac Asimov",
		Title:  "the shark",
		Score:  4,
		Price:  15.50,
	}

	book4 := Book{
		Author: "Dog",
		Title:  "the long bark",
		Score:  5,
		Price:  8.99,
	}

	s.AddBook(book1)
	s.AddBook(book2)
	s.AddBook(book3)
	s.AddBook(book4)
}

func (s *Store) AddBook(b Book) {
	key := strings.ToLower(b.Title)
	s.Books[key] = b

}

func (s *Store) ListBooksAsJson(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.Books)
}

// Could do this without using gorilla/mux, but not as cleanly.
func (s *Store) GetBookByTitle(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	key := strings.ToLower(vars["title"])

	if book, exists := s.Books[key]; exists {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(book)
		return
	}

	closeMatches := make(map[int][]Book)
	bestMatches := make([]Book, 0, 2)
	scores := make([]int, 0)
	seenScores := make(map[int]bool)
	threshold := 70

	// This is obviously not the most efficient way to do this, but it is simple and works for small datasets. Could use
	// goroutines to compute in parallel as well. For larger
	// datasets you would want to use something like elasticsearch.
	for title, book := range s.Books {
		score := fuzzy.Ratio(key, title)
		if score < threshold {
			continue
		}

		closeMatches[score] = append(closeMatches[score], book)

		if !seenScores[score] {

			seenScores[score] = true
			scores = append(scores, score)

		}

	}
	if len(closeMatches) == 0 {
		http.Error(w, "No books found matching the title", http.StatusNotFound)
		return

	}
	// Sort scores in descending order, if scores are equal, sort by title. This ensures consistent ordering
	// accross different runs if there are ties in score.

	sort.Slice(scores, func(i, j int) bool {

		return scores[i] > scores[j]
	})

	for _, score := range scores {

		sort.Slice((closeMatches[score]), func(i, j int) bool {
			return closeMatches[score][i].Title < closeMatches[score][j].Title
		})

	}

	// Extract the top 2 scoring books
outer:
	for _, score := range scores {
		for _, book := range closeMatches[score] {

			bestMatches = append(bestMatches, book)

			if len(bestMatches) >= 2 {
				break outer
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bestMatches)

}
