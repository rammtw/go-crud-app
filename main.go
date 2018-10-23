package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main(){
	handler := http.NewServeMux()

	handler.HandleFunc("/hello", Logger(helloHandler))

	handler.HandleFunc("/book/", Logger(bookHandler))

	handler.HandleFunc("/books/", Logger(booksHandler))

	s := http.Server{
		Addr: ":8081",
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

type Resp struct {
	Message string
	Error string
}

type Book struct {
	Id string `json:"id"`
	Author string `json:"author"`
	Name string `json:"name"`
}

type BookStore struct {
	books []Book
}

var bookStore = BookStore {
	books: make([]Book, 0),
}

func helloHandler(w http.ResponseWriter, r *http.Request){
	resp := Resp {
		Message: "hello",
	}

	respJson, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)

	w.Write(respJson)
}

func bookHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		handleGetBook(w, r)
	} else if r.Method == http.MethodPost {
		handleAddBook(w, r)
	} else if r.Method == http.MethodDelete {
		handleDeleteBook(w, r)
	} else if r.Method == http.MethodPut {
		handleUpdateBook(w, r)
	} else {
		//handleDefault(w, r)
	}
}

func booksHandler(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)

	booksJson, _ := json.Marshal(bookStore.GetBooks())

	w.Write(booksJson)
}


func handleGetBook(w http.ResponseWriter, r *http.Request){
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	book := bookStore.FindBookById(id)

	if book == nil {
		w.WriteHeader(http.StatusNotFound)

		respJson, _ := json.Marshal(fmt.Sprintf("Book with id %s not found", id))

		w.Write(respJson)

		return
	}

	w.WriteHeader(http.StatusOK)

	respJson, _ := json.Marshal(book)

	w.Write(respJson)
}

func handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	var resp Resp

	err := bookStore.DeleteBook(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		resp.Error = err.Error()

		respJson, _ := json.Marshal(resp)

		w.Write(respJson)

		return
	}

	booksHandler(w, r)
}

func handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	decoder := json.NewDecoder(r.Body)

	var book Book

	err := decoder.Decode(&book)

	log.Printf("handleUpdateBook is run")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("Bad request. %v", err))

		w.Write(error)

		return
	}

	book.Id = id

	err = bookStore.UpdateBook(book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("%v", err))

		w.Write(error)

		return
	}

	handleGetBook(w, r)
}

func handleAddBook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var book Book

	err := decoder.Decode(&book)

	log.Printf("handleAddBook is run")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("Bad request. %v", err))

		w.Write(error)

		return
	}

	err = bookStore.AddBook(book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("Bad request. %v", err))

		w.Write(error)

		return
	}

	booksHandler(w, r)
}

func (s BookStore) FindBookById(id string) *Book{
	for _, book := range s.books{
		if book.Id == id {
			return &book
		}
	}

	return nil
}

func (s BookStore) GetBooks() []Book{
	return s.books
}

func (s *BookStore) UpdateBook(book Book) error {
	for i, bk := range s.books{
		if bk.Id == book.Id {
			s.books[i] = book
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Book with id %s not found", book.Id))
}

func (s *BookStore) DeleteBook(id string) error {
	for i, bk := range s.books {
		if bk.Id == id {
			s.books = append(s.books[:i], s.books[i+1:]...)
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Book with id %s not found", id))
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log.Printf("server [net/http] method [%s] connection from [%v]", r.Method, r.RemoteAddr)

		next.ServeHTTP(w, r)
	}
}

func (s *BookStore) AddBook(book Book) error{
	bk := s.FindBookById(book.Id)
	if bk != nil {
		return errors.New(fmt.Sprintf("Book with id %s already exists", book.Id))

	}
	s.books = append(s.books, book)

	return nil
}