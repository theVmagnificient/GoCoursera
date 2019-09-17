package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	///	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)


type XmlUser struct {
	Id 			int `xml:"id"`
	FirstName 	string `xml:"first_name"`
	LastName 	string `xml:"last_name"`
	Age      	int `xml:"age"`
	About       string `xml:"about"`
	Gender      string `xml:"gender"`
}

type Users struct {
	Version string `xml:"version,attr"`
	List    []XmlUser `xml:"row"`
}

func GetData(path string) (Users, error) {
	xmlFile, err := os.Open(path)
	var users Users

	if err != nil {
		fmt.Printf("error: %v", err)
		return users, err
	}
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)


	fmt.Println("Successfully Opened xml")

	err = xml.Unmarshal(byteValue, &users)

	if err != nil {
		//fmt.Println("Error unmarshalling json: %v", err)
		return users, err
	}

	return users, nil
}

func GetRespJson(user XmlUser) User {
	var JsonUser User

	JsonUser.Name = user.FirstName + user.LastName
	JsonUser.Id = user.Id
	JsonUser.About = user.About
	JsonUser.Age = user.Age
	JsonUser.Gender = user.Gender

	return JsonUser
}

func CheckQueryForUser(user XmlUser, q string) bool {
	if q == "" {
		return true
	}
	return user.FirstName == q || user.LastName == q || strings.Contains(user.About, q)
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("AccessToken")

	if token != "12345" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	orderField := r.FormValue("order_field")
	switch orderField {
	case "timeout":
		time.Sleep(5 * time.Second)
	case "broken_json":
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"status": 400`) //broken json
		if err != nil {
			return
		}
		return
	case "internal_error":
		w.WriteHeader(http.StatusInternalServerError)
		return
	default:
		break
	}

	if orderField == "" {
		orderField = "name"
	}
	if orderField != "name" && orderField != "age" && orderField != "id" {
		if orderField != "error_json" && orderField != "unk_bad_req" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			respErr, err := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
			if err != nil {
				//fmt.Println("Error marshalling json %v", err)
				return
			}
			_, err = w.Write(respErr)

			if err != nil {
				//fmt.Println("Error writing to header %v", err)
				return
			}
			return
		} else if orderField == "error_json"{
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(w, `"{}}{{}{}{}}{status": 400, smkasmdkasmdkasmdkamkdmakmd`)
			if err != nil {

				return
			}
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(w, `{"Error": "SHIT HAPPENED"} `)
			if err != nil {

				return
			}
			return
		}
	}

	users, err := GetData("dataset.xml")

	if err != nil {
		fmt.Println(err)
		return
	}

	query := r.FormValue("query")

	var respUsers []User
	for i := 0; i < len(users.List); i++ {
		if ok := CheckQueryForUser(users.List[i], query); ok {
			respUsers = append(respUsers, GetRespJson(users.List[i]))
		}
	}

	if orderField == "name" || orderField == "" {
		sort.Slice(respUsers, func(i, j int) bool {
			return respUsers[i].Name > respUsers[j].Name
		})
	} else if orderField == "age" {
		sort.Slice(respUsers, func(i, j int) bool {
			return respUsers[i].Age > respUsers[j].Age
		})
	} else if orderField == "id" {
		sort.Slice(respUsers, func(i, j int) bool {
			return respUsers[i].Id > respUsers[j].Id
		})
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	js, err := json.Marshal(respUsers)

	if err != nil {
		//fmt.Println("Error marshalling json %v", err)
		return
	}
	_, err = w.Write(js)

	if err != nil {
		//fmt.Println("Error writing to header %v", err)
		return
	}
}

type TestCase struct {
	 Limit        int
	 Offset       int
	 Query        string
	 OrderField   string
	 OrderBy      int
	 IsError      bool
}
func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	s := SearchClient{
		AccessToken: "12345",
		URL:         ts.URL,
	}

	s1 := SearchClient{
		AccessToken: "1",
		URL:         ts.URL,
	}

	s2 := SearchClient{
		AccessToken: "12345",
		URL:         "qwert",
	}

	tests := []TestCase {
		{
			Limit:      0,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "name",
			OrderBy:    0,
			IsError:    false,
	    },
		{
			Limit:      0,
			Offset:     0,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
			IsError:    false,
		},
		{
			Limit:      0,
			Offset:     0,
			Query:      "",
			OrderField: "qweqwr",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      -1,
			Offset:     0,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      0,
			Offset:     -1,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      0,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "timeout",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      0,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "broken_json",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      0,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "error_json",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      0,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "unk_bad_req",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      1,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "name",
			OrderBy:    0,
			IsError:    false,
		},
		{
			Limit:      1,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "internal_error",
			OrderBy:    0,
			IsError:    true,
		},
		{
			Limit:      26,
			Offset:     0,
			Query:      "Dillard",
			OrderField: "name",
			OrderBy:    0,
			IsError:    false,
		},
	}

	users, _ := GetData("dataset.xml")

	var jUsers []User
	for i := 0; i < len(users.List); i++ {
			jUsers = append(jUsers, GetRespJson(users.List[i]))
	}


	for caseNum, test := range tests {
		var tmpUsers []User

		req := SearchRequest{
			test.Limit,
			test.Offset,
			test.Query,
			test.OrderField,
			test.OrderBy,
		}

		resp, err := s.FindUsers(req)

		if err != nil && !test.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
			continue
		}
		if err == nil && test.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
			continue
		}

		for i := 0; i < len(users.List); i++ {
			if ok := CheckQueryForUser(users.List[i], test.Query); ok {
				tmpUsers = append(tmpUsers, GetRespJson(users.List[i]))
			}
		}

		if test.OrderField == "name" || test.OrderField == "" {
			sort.Slice(tmpUsers, func(i, j int) bool {
				return tmpUsers[i].Name > tmpUsers[j].Name
			})
		} else if test.OrderField == "age" {
			sort.Slice(tmpUsers, func(i, j int) bool {
				return tmpUsers[i].Age > tmpUsers[j].Age
			})
		} else if test.OrderField == "id" {
			sort.Slice(tmpUsers, func(i, j int) bool {
				return tmpUsers[i].Id > tmpUsers[j].Id
			})
		}
		if err == nil  {
			if len(resp.Users) != len(tmpUsers) && test.Limit != len(resp.Users){
				t.Errorf("[%d] wrong len, expected %#v, got %#v", caseNum, len(tmpUsers), len(resp.Users))

				for i := 0; i < len(resp.Users); i++ {
					if resp.Users[i] != tmpUsers[i] {
						t.Errorf("Data missmatch in position [%d]", i)
					}
				}
			}

		}
	}
	_, err := s1.FindUsers(SearchRequest{tests[0].Limit, tests[0].Offset, tests[0].Query, tests[0].OrderField, tests[0].OrderBy})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	_, err = s2.FindUsers(SearchRequest{tests[0].Limit, tests[0].Offset, tests[0].Query, tests[0].OrderField, tests[0].OrderBy})

	if err == nil {
		t.Errorf("Expected error, got nil")
	}


	ts.Close()
}


//func main() {
//	TestSearch()
//}