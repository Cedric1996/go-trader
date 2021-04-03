/*
 * @Author: cedric.jia
 * @Date: 2021-03-18 15:47:59
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-18 16:07:46
 */
package app

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Test struct {
	Test string `json:"test"`
}

type CreateRepoResposne struct {
	User          string `json:"plastic_user"`
	ClientMachine string `json:"PLASTIC_CLIENTMACHINE"`
	Server        string `json:"PLASTIC_SERVER"`
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
	fmt.Println("hello")
}

func post(w http.ResponseWriter, req *http.Request) {
	result := &CreateRepoResposne{}
	err := json.NewDecoder(req.Body).Decode(&result)
	if err != nil {
		fmt.Errorf("get POST with error: %s", err)
		return
	}
	fmt.Printf("Get POST: %s %s %s\n", result.User, result.ClientMachine, result.Server)

	// fmt.Printf("Get POST: %s %s %s\n", result.user, result.clientMachine, result.server)
}

func RunServer() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/post", post)
	http.ListenAndServe(":8090", nil)
}
