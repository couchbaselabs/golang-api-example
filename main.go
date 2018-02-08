package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	gocb "gopkg.in/couchbase/gocb.v1"
)

var bucket *gocb.Bucket

type CreditCard struct {
	Brand      string `json:"brand"`
	Number     string `json:"number"`
	Expiration string `json:"expiration"`
}

type Customer struct {
	Id          string       `json:"id,omitempty"`
	Type        string       `json:"type"`
	Firstname   string       `json:"firstname"`
	Lastname    string       `json:"lastname"`
	CreditCards []CreditCard `json:"creditcards"`
}

type Product struct {
	Id    string  `json:"id,omitempty"`
	Type  string  `json:"type"`
	Name  string  `json:"name"`
	Price float32 `json:"price"`
}

type Receipt struct {
	Id       string    `json:"id,omitempty"`
	Type     string    `json:"type"`
	Customer Customer  `json:"customer"`
	Products []Product `json:"products"`
	Total    float32   `json:"total"`
}

func RootEndpoint(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(200)
	response.Write([]byte("Hello World"))
}

func GetCustomersEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var customers []Customer
	query := gocb.NewN1qlQuery("SELECT META().id, " + bucket.Name() + ".* FROM " + bucket.Name() + " WHERE type = 'customer'")
	rows, err := bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var row Customer
	for rows.Next(&row) {
		customers = append(customers, row)
	}
	json.NewEncoder(response).Encode(customers)
}

func CreateCustomerEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var customer Customer
	json.NewDecoder(request.Body).Decode(&customer)
	customer.Type = "customer"
	id := uuid.NewV4().String()
	_, err := bucket.Insert(id, customer, 0)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	response.Write([]byte(`{ "id": "` + id + `" }`))
}

func GetCustomerEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	routeParams := mux.Vars(request)
	var customer Customer
	customer.Id = routeParams["id"]
	_, err := bucket.Get(routeParams["id"], &customer)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(customer)
}

func AddCreditCardEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	routeParams := mux.Vars(request)
	var creditcard CreditCard
	json.NewDecoder(request.Body).Decode(&creditcard)
	_, err := bucket.MutateIn(routeParams["id"], 0, 0).ArrayAppend("creditcards", creditcard, true).Execute()
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(creditcard)
}

func GetCreditCardsForCustomerEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	routeParams := mux.Vars(request)
	fragment, err := bucket.LookupIn(routeParams["id"]).Get("creditcards").Execute()
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var creditcards []CreditCard
	fragment.Content("creditcards", &creditcards)
	json.NewEncoder(response).Encode(creditcards)
}

func GetProductsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var products []Product
	query := gocb.NewN1qlQuery("SELECT META().id, " + bucket.Name() + ".* FROM " + bucket.Name() + " WHERE type = 'product'")
	rows, err := bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var row Product
	for rows.Next(&row) {
		products = append(products, row)
	}
	json.NewEncoder(response).Encode(products)
}

func CreateProductEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var product Product
	json.NewDecoder(request.Body).Decode(&product)
	product.Type = "product"
	id := uuid.NewV4().String()
	_, err := bucket.Insert(id, product, 0)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	response.Write([]byte(`{ "id": "` + id + `" }`))
}

func GetReceiptsForCustomerEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	routeParams := mux.Vars(request)
	query := gocb.NewN1qlQuery("SELECT META(receipts).id, receipts.* FROM " + bucket.Name() + " AS receipts WHERE receipts.type = 'receipt' AND receipts.customer.id = $1")
	var params []interface{}
	params = append(params, routeParams["id"])
	rows, err := bucket.ExecuteN1qlQuery(query, params)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var receipts []Receipt
	var row Receipt
	for rows.Next(&row) {
		receipts = append(receipts, row)
	}
	json.NewEncoder(response).Encode(receipts)
}

func CreateReceiptEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var receipt Receipt
	json.NewDecoder(request.Body).Decode(&receipt)
	productsIds := make([]string, len(receipt.Products))
	for i, product := range receipt.Products {
		productsIds[i] = product.Id
	}
	query := gocb.NewN1qlQuery("SELECT (SELECT VALUE { META(customer).id, customer.firstname, customer.lastname, customer.type } FROM " + bucket.Name() + " AS customer USE KEYS $2)[0] AS customer, products FROM " + bucket.Name() + " AS receipt USE KEYS $2 LET products = (SELECT META(product).id, product.name, product.price, product.type FROM " + bucket.Name() + " AS product USE KEYS $1)")
	var params []interface{}
	params = append(params, productsIds)
	params = append(params, receipt.Customer.Id)
	rows, err := bucket.ExecuteN1qlQuery(query, params)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	receipt = Receipt{}
	err = rows.One(&receipt)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	receipt.Type = "receipt"
	for _, product := range receipt.Products {
		receipt.Total += product.Price
	}
	id := uuid.NewV4().String()
	bucket.Insert(id, receipt, 0)
	response.Write([]byte(`{ "id": "` + id + `" }`))
}

func main() {
	fmt.Println("Starting the application...")
	cluster, _ := gocb.Connect("couchbase://192.168.1.31")
	cluster.Authenticate(gocb.PasswordAuthenticator{Username: "demo", Password: "password"})
	bucket, _ = cluster.OpenBucket("demo", "")
	router := mux.NewRouter()
	router.HandleFunc("/", RootEndpoint).Methods("GET")
	router.HandleFunc("/customers", GetCustomersEndpoint).Methods("GET")
	router.HandleFunc("/customer", CreateCustomerEndpoint).Methods("POST")
	router.HandleFunc("/customer/{id}", GetCustomerEndpoint).Methods("GET")
	router.HandleFunc("/customer/orders/{id}", GetReceiptsForCustomerEndpoint).Methods("GET")
	router.HandleFunc("/products", GetProductsEndpoint).Methods("GET")
	router.HandleFunc("/product", CreateProductEndpoint).Methods("POST")
	router.HandleFunc("/order", CreateReceiptEndpoint).Methods("POST")
	router.HandleFunc("/customer/creditcard/{id}", AddCreditCardEndpoint).Methods("PUT")
	router.HandleFunc("/customer/creditcard/{id}", GetCreditCardsForCustomerEndpoint).Methods("GET")
	fmt.Println("Listening at :12345")
	log.Fatal(http.ListenAndServe(":12345", router))
}
