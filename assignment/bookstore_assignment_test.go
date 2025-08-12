package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	store := NewStore()
	require.NotNil(t, store)
}

func TestAddBook(t *testing.T) {
	store := NewStore()
	store.AddBook(Book{
		Title:  "Normandy manual",
		Author: "commander shepard",
	})

	require.Len(t, store.Books, 1)
	require.Equal(t, store.Books["normandy manual"].Author, "commander shepard")
}

func TestFillupStore(t *testing.T) {
	store := NewStore()
	store.fillupStore()

	require.Len(t, store.Books, 4)
}

func TestListBooksAsJson(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:3000/books", nil)
	w := httptest.NewRecorder()

	store := NewStore()
	store.fillupStore()
	store.ListBooksAsJson(w, req)

	response := w.Result()
	require.Equal(t, http.StatusOK, response.StatusCode)

	books := unmarshalResponseBodyToBooks(t, response)
	require.Len(t, books, 4)

	expected := Book{
		Author: "Stephen King",
		Title:  "the long walk",
		Score:  2,
		Price:  9.99,
	}

	require.Contains(t, books, expected)
}

// added a fmt.Println(string(data)) to inspect the structure of the response body.
// The problem was that json.unmarshal(data, &books) expected an json array but got an json object instead.
// Solution was to create map variable to accomodate the response structure.
// Since I didn't want to change the return signature of the function, I then looped through this
// map to create a slice of books.
func unmarshalResponseBodyToBooks(t *testing.T, res *http.Response) []Book {
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	//fmt.Println(string(data))
	var books []Book
	err = json.Unmarshal(data, &books)
	if err == nil {
		return books
	}
	// If the above unmarshal fails, we assume the response is a map
	var booksMap map[string]Book

	err = json.Unmarshal(data, &booksMap)
	require.NoError(t, err)

	for _, book := range booksMap {
		books = append(books, book)
	}

	return books

}

func TestGetBookByTitle(t *testing.T) {

	// Test for a book that exists with exact title match, ignoring case differences

	req1 := httptest.NewRequest("GET", "http://localhost:3000/books/THE%20LoNg%20WAlk", nil)
	w1 := httptest.NewRecorder()

	store := NewStore()
	store.fillupStore()

	router := mux.NewRouter()
	router.HandleFunc("/books/{title}", store.GetBookByTitle)

	// This ensures the /mux context is correctly set
	router.ServeHTTP(w1, req1)

	response1 := w1.Result()
	require.Equal(t, http.StatusOK, response1.StatusCode)

	book := unmarshalResponseBodyToBook(t, response1)

	expected := Book{
		Author: "Stephen King",
		Title:  "the long walk",
		Score:  2,
		Price:  9.99,
	}

	require.Equal(t, book, expected)

	// Test for a book that does not exist with no close matches

	req2 := httptest.NewRequest("GET", "http://localhost:3000/books/NotFound", nil)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)
	response2 := w2.Result()

	require.Equal(t, http.StatusNotFound, response2.StatusCode)

	// Test for a book that doesn't exist but with close matches

	req3 := httptest.NewRequest("GET", "http://localhost:3000/books/The%20long%20shark", nil)
	w3 := httptest.NewRecorder()

	router.ServeHTTP(w3, req3)
	response3 := w3.Result()

	require.Equal(t, http.StatusOK, response3.StatusCode)

	books := unmarshalResponseBodyToBooks(t, response3)

	expectedBooks := []Book{

		{
			Author: "Dog",
			Title:  "the long bark",
			Score:  5,
			Price:  8.99,
		},

		{
			Author: "Stephen King",
			Title:  "the long walk",
			Score:  2,
			Price:  9.99,
		},
	}

	require.Equal(t, books, expectedBooks)

}

func unmarshalResponseBodyToBook(t *testing.T, res *http.Response) Book {
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var book Book
	err = json.Unmarshal(data, &book)
	require.NoError(t, err)

	return book
}
