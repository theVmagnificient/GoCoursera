package main
import (
"encoding/json"
"strings"
"strconv"
"fmt"
"net/http"
)


func postCheck(r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	return false
}



func checkAuth(r *http.Request) bool {
	if code := r.Header.Get("X-Auth"); code == "100500" {
		return true
	}
	return false
}

type FR map[string]interface{}
func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"error\":\"unknown method\"}")
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"error\":\"unknown method\"}")
	}
}


// Profile http handler	
func (s *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if false {
		if isPost := postCheck(r); isPost != true {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Fprintln(w, "{\"error\":\"bad method\"}")
			return
		}
	}
	if false {
		if isAuth := checkAuth(r); !isAuth {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "{\"error\":\"unauthorized\"}")
			return
		}
	}
	//params := ValProfileParamsProfileParams(w, r)
	params := ValProfileParams(w, r)

	if params != nil {
		resp, err := s.Profile(ctx, *params)
	
		if err != nil {
			switch err.(type) {
				default:
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "{\"error\":\"" + err.Error() + "\"}")
				case ApiError:
					w.WriteHeader(err.(ApiError).HTTPStatus)
					fmt.Fprintf(w, "{\"error\":\"" + err.Error() + "\"}")
 			}
		} else {
			fullResp := FR{"error": "", "response": resp}
			jsonData, err := json.Marshal(fullResp)
			if err != nil {
				fmt.Println(err)
			}
			w.Write(jsonData)
		}
	}
}
func ValProfileParams(w http.ResponseWriter, r *http.Request) *ProfileParams {
	var params ProfileParams
	// paramsFillString_login
	params.Login = r.FormValue("login")

	// paramsValidateStrRequired_ProfileParams.login
	if params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"login must me not empty\"}")
		return nil
	}

	// paramsFillString_status
	params.Status = r.FormValue("status")

	// paramDefaultStr_ProfileParams.status
	if params.Status == "" {
		params.Status = "user"
	}
	
	// paramsValidateStrEnum_ProfileParams.status
	var enums string

	 
	enums += "user, "
	 
	enums += "moderator, "
	 
	enums += "admin, "
	
	enums = strings.TrimSuffix(enums, ", ")
	list := strings.Split(enums, ", ")
	flag := false

	for i := 0; i < len(list); i++ {
		if list[i] == params.Status {
			flag = true
		}
	}

	if !flag  {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"status must be one of [" + enums + "]\"}")
		return nil
	}

 return &params
} // End of validator for  ProfileParams

// Create http handler	
func (s *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if true {
		if isPost := postCheck(r); isPost != true {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Fprintln(w, "{\"error\":\"bad method\"}")
			return
		}
	}
	if true {
		if isAuth := checkAuth(r); !isAuth {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "{\"error\":\"unauthorized\"}")
			return
		}
	}
	//params := ValCreateParamsCreateParams(w, r)
	params := ValCreateParams(w, r)

	if params != nil {
		resp, err := s.Create(ctx, *params)
	
		if err != nil {
			switch err.(type) {
				default:
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "{\"error\":\"" + err.Error() + "\"}")
				case ApiError:
					w.WriteHeader(err.(ApiError).HTTPStatus)
					fmt.Fprintf(w, "{\"error\":\"" + err.Error() + "\"}")
 			}
		} else {
			fullResp := FR{"error": "", "response": resp}
			jsonData, err := json.Marshal(fullResp)
			if err != nil {
				fmt.Println(err)
			}
			w.Write(jsonData)
		}
	}
}
func ValCreateParams(w http.ResponseWriter, r *http.Request) *CreateParams {
	var params CreateParams
	// paramsFillString_login
	params.Login = r.FormValue("login")

	// paramsValidateStrRequired_CreateParams.login
	if params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"login must me not empty\"}")
		return nil
	}

	if len(params.Login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"login len must be >= 10\"}")
		return nil
	}
	
	// paramsFillString_full_name
	params.Name = r.FormValue("full_name")

	// paramsFillString_status
	params.Status = r.FormValue("status")

	// paramDefaultStr_CreateParams.status
	if params.Status == "" {
		params.Status = "user"
	}
	
	// paramsValidateStrEnum_CreateParams.status
	var enums string

	 
	enums += "user, "
	 
	enums += "moderator, "
	 
	enums += "admin, "
	
	enums = strings.TrimSuffix(enums, ", ")
	list := strings.Split(enums, ", ")
	flag := false

	for i := 0; i < len(list); i++ {
		if list[i] == params.Status {
			flag = true
		}
	}

	if !flag  {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"status must be one of [" + enums + "]\"}")
		return nil
	}

	// paramsFillInt_age
	if tmp, err := strconv.Atoi(r.FormValue("age")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"age must be int\"}")
		return nil
	} else {
		params.Age = tmp
	}

	if params.Age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"age must be >= 0\"}")
		return nil
	}
	
	// paramLenCheck_CreateParams.age
	if params.Age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"age must be <= 128\"}")
		return nil
	}
 return &params
} // End of validator for  CreateParams

// Create http handler	
func (s *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if true {
		if isPost := postCheck(r); isPost != true {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Fprintln(w, "{\"error\":\"bad method\"}")
			return
		}
	}
	if true {
		if isAuth := checkAuth(r); !isAuth {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "{\"error\":\"unauthorized\"}")
			return
		}
	}
	//params := ValCreateParamsOtherCreateParams(w, r)
	params := ValOtherCreateParams(w, r)

	if params != nil {
		resp, err := s.Create(ctx, *params)
	
		if err != nil {
			switch err.(type) {
				default:
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "{\"error\":\"" + err.Error() + "\"}")
				case ApiError:
					w.WriteHeader(err.(ApiError).HTTPStatus)
					fmt.Fprintf(w, "{\"error\":\"" + err.Error() + "\"}")
 			}
		} else {
			fullResp := FR{"error": "", "response": resp}
			jsonData, err := json.Marshal(fullResp)
			if err != nil {
				fmt.Println(err)
			}
			w.Write(jsonData)
		}
	}
}
func ValOtherCreateParams(w http.ResponseWriter, r *http.Request) *OtherCreateParams {
	var params OtherCreateParams
	// paramsFillString_username
	params.Username = r.FormValue("username")

	// paramsValidateStrRequired_OtherCreateParams.username
	if params.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"username must me not empty\"}")
		return nil
	}

	if len(params.Username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"username len must be >= 3\"}")
		return nil
	}
	
	// paramsFillString_account_name
	params.Name = r.FormValue("account_name")

	// paramsFillString_class
	params.Class = r.FormValue("class")

	// paramDefaultStr_OtherCreateParams.class
	if params.Class == "" {
		params.Class = "warrior"
	}
	
	// paramsValidateStrEnum_OtherCreateParams.class
	var enums string

	 
	enums += "warrior, "
	 
	enums += "sorcerer, "
	 
	enums += "rouge, "
	
	enums = strings.TrimSuffix(enums, ", ")
	list := strings.Split(enums, ", ")
	flag := false

	for i := 0; i < len(list); i++ {
		if list[i] == params.Class {
			flag = true
		}
	}

	if !flag  {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"class must be one of [" + enums + "]\"}")
		return nil
	}

	// paramsFillInt_level
	if tmp, err := strconv.Atoi(r.FormValue("level")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"level must be int\"}")
		return nil
	} else {
		params.Level = tmp
	}

	if params.Level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"level must be >= 1\"}")
		return nil
	}
	
	// paramLenCheck_OtherCreateParams.level
	if params.Level > 50 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"level must be <= 50\"}")
		return nil
	}
 return &params
} // End of validator for  OtherCreateParams
