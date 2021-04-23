package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/weirdwiz/online_judge/authentication/internal/app/dbclient"
	"github.com/weirdwiz/online_judge/authentication/internal/app/model"
)

var mySigningKey = []byte("signingKey")

type JWTstring struct {
	Token string `json:"token"`
}

func GenerateJWT(email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 60).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {

			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {

			fmt.Fprintf(w, "Not Authorized")
		}
	})
}

func HandleRegistration(w http.ResponseWriter, r *http.Request) {
	var user model.User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Errorf("error decoding user"))
	}
	success, err := DBClient.CreateUser(user)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}
	fmt.Fprintf(w, "Status: %t", success)
}

func WriteError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, err.Error())
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if r.Header.Get("Content-Type") == "application/json" {
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			WriteError(w, http.StatusBadRequest, fmt.Errorf("error decoding user"))
		}
	} else {
		email := r.FormValue("email")
		password := r.FormValue("password")
		accounttype := r.FormValue("accounttype")

		user.Email = email
		user.Password = password
		user.AccountType = accounttype
	}

	_, err := DBClient.Login(user.Email, user.Password, user.AccountType)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}
	tokenString, err := GenerateJWT(user.Email)
	if err != nil {
		fmt.Fprintf(w, "Error generating token string")
	}
	token := &JWTstring{
		Token: tokenString,
	}
	data, _ := json.Marshal(token)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

// func HandleAddBatch(w http.ResponseWriter, r *http.Request) {
// 	var batch model.Batch
// 	if r.Header.Get("Content-Type") == "application/json" {
// 		err := json.NewDecoder(r.Body).Decode(&batch)
// 		if err != nil {
// 			WriteError(w, http.StatusBadRequest, fmt.Errorf("Error Decoding Batch"))
// 		}
// 	} else {
// 		name := r.FormValue("name")
// 		students := r.FormValue("students")

// 		batch.Name = name
// 		batch.Students = students
// 	}

// 	_, err := DBClient.AddBatch(batch.Name, batch.Students)
// 	if err != nil {
// 		WriteError(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
// 	w.WriteHeader(http.StatusOK)
// 	w.Write(data)

// }

var DBClient dbclient.IDBClient
