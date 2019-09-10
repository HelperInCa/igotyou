package main

import (
	elastic "gopkg.in/olivere/elastic.v3"

	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	TYPE_USER = "user"
)

// match one or more words: lower case | num | underline
var (
	usernamePattern = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
}

// checks whether a pair of username and password is stored in ES.
func checkUser(username, password string) bool {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		fmt.Printf("ES is not setup %v\n", err)
		return false
	}

	// Search with a term query
	termQuery := elastic.NewTermQuery("username", username)
	queryResult, err := es_client.Search().
		Index(INDEX).
		Query(termQuery).
		Pretty(true).
		Do()
	if err != nil {
		fmt.Printf("ES query failed %v\n", err)
		return false
	}

	var tyu User
	for _, item := range queryResult.Each(reflect.TypeOf(tyu)) {
		u := item.(User)
		return u.Password == password && u.Username == username
	}
	// If no user exists, return false.
	return false
}

// Add a new user. Return true if successfully.
func addUser(user User) bool {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		fmt.Printf("ES is not setup %v\n", err)
		return false
	}

	termQuery := elastic.NewTermQuery("username", user.Username)
	queryResult, err := es_client.Search().
		Index(INDEX).
		Query(termQuery).
		Pretty(true).
		Do()
	if err != nil {
		fmt.Printf("ES query failed %v\n", err)
		return false
	}

	if queryResult.TotalHits() > 0 {
		fmt.Printf("User %s already exists, cannot create duplicate user.\n", user.Username)
		return false
	}

	_, err = es_client.Index().
		Index(INDEX).
		Type(TYPE_USER).
		Id(user.Username).
		BodyJson(user).
		Refresh(true).
		Do()
	if err != nil {
		fmt.Printf("ES save user failed %v\n", err)
		return false
	}

	return true
}

/*
1. Decode a user from request (POST)
2. Check whether username and password are empty, if any of them is empty,
   call http.Error(w, "Empty password or username", http.StatusInternalServerError)
3. Otherwise, call addUser, if true, return a message “User added successfully”
4. If else, call http.Error(w, "Failed to add a new user", http.StatusInternalServerError)
5. Set header to be w.Header().Set("Content-Type", "text/plain")
   w.Header().Set("Access-Control-Allow-Origin", "*")
*/
// If signup is successful, a new session is created.
func signupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one signup request")

	decoder := json.NewDecoder(r.Body)
	var u User
	if err := decoder.Decode(&u); err != nil {
		panic(err)
		return
	}

	if u.Username != "" && u.Password != "" && usernamePattern(u.Username) {
		if addUser(u) {
			fmt.Println("User added successfully.")
			w.Write([]byte("User added successfully."))
		} else {
			fmt.Println("Failed to add a new user.")
			http.Error(w, "Failed to add a new user", http.StatusInternalServerError) //500
		}
	} else {
		fmt.Println("Empty password or username.")
		http.Error(w, "Empty password or username", http.StatusInternalServerError) //500
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

/*
1.
2.
3. Create a new token
*/
// If login is successful, a new token is created.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one login request")
	// Decode a user from request(POST)
	decoder := json.NewDecoder(r.Body)
	var u User
	if err := decoder.Decode(&u); err != nil {
		panic(err)
		return
	}
	// Make sure the user's credential is correct
	if checkUser(u.Username, u.Password) {
		// Create a new token object to store
		token := jwt.New(jwt.SigningMethodHS256)
		// Convert it into a map for lookup
		claims := token.Claims.(jwt.MapClaims)
		/*
			Set token claims
			Store username and expiration into it.
		*/
		claims["username"] = u.Username
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		/* Sign the token with our secret */
		tokenString, _ := token.SignedString(mySigningKey)

		/* Finally, write the token to the browser window */
		w.Write([]byte(tokenString))
	} else {
		fmt.Println("Invalid password or username.")
		http.Error(w, "Invalid password or username", http.StatusForbidden) //403
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
