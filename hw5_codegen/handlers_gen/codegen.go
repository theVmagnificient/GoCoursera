package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	//"strings"
)


var (
	reRequred   = regexp.MustCompile(",(required),")
	reParamname = regexp.MustCompile(",paramname=([^,]*)")
	reEnum      = regexp.MustCompile(",enum=([^,]*)") //TODO strings.Split(..,"|")
	reDefault   = regexp.MustCompile(",default=([^,]*)")
	reMin       = regexp.MustCompile(",min=([^,]*)")
	reMax       = regexp.MustCompile(",max=([^,]*)")
)
type tpl struct {
	FieldName string
}

type httpTpl struct {
	StructName string
	FuncName   string
	IsPost     string
	CheckAuth  string
	CheckPost  string
	ApiName    string
}

type descr struct {
	URL 	string `json:"url"`
	Auth    bool   `json:"auth"`
	Method  string `json:"method"`
}

type fillTpl struct {
	StructType  string
	ParamName	string
	FieldName	string
	EnumVals    []string
	DefValStr   string
	DefValInt	int
	MinVal      int
	MaxVal      int
}


var (
	intTpl = template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	var {{.FieldName}}Raw uint32
	binary.Read(r, binary.LittleEndian, &{{.FieldName}}Raw)
	in.{{.FieldName}} = int({{.FieldName}}Raw)
`))

	strTpl = template.Must(template.New("strTpl").Parse(`
	// {{.FieldName}}
	var {{.FieldName}}LenRaw uint32
	binary.Read(r, binary.LittleEndian, &{{.FieldName}}LenRaw)
	{{.FieldName}}Raw := make([]byte, {{.FieldName}}LenRaw)
	binary.Read(r, binary.LittleEndian, &{{.FieldName}}Raw)
	in.{{.FieldName}} = string({{.FieldName}}Raw)
`))


	postCheck = `
func postCheck(r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	return false
}
`

	authCheck = `
func checkAuth(r *http.Request) bool {
	if code := r.Header.Get("X-Auth"); code == "100500" {
		return true
	}
	return false
}
`
	strFillTpl = template.Must(template.New("strFillTpl").Parse(`
	// paramsFillString_{{.FieldName}}
	params.{{.ParamName}} = r.FormValue("{{.FieldName}}")
`))

	intFillTpl = template.Must(template.New("intFillTpl").Parse(`
	// paramsFillInt_{{.FieldName}}
	if tmp, err := strconv.Atoi(r.FormValue("{{.FieldName}}")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} must be int\"}")
		return nil
	} else {
		params.{{.ParamName}} = tmp
	}
`))
	avStrRequiredTpl = template.Must(template.New("avStrRequiredTpl").Parse(`
	// paramsValidateStrRequired_{{.StructType}}.{{.FieldName}}
	if params.{{.ParamName}} == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} must me not empty\"}")
		return nil
	}
`))
	avIntRequiredTpl = template.Must(template.New("avIntRequiredTpl").Parse(`
	// paramsValidateIntRequired_{{.StructType}}.{{.FieldName}}
	if params.{{.ParamName}} == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} must be not empty\"}")
		return nil
	}
`))

	avStrEnumTpl = template.Must(template.New("avStrEnumTpl").Parse(`
	// paramsValidateStrEnum_{{.StructType}}.{{.FieldName}}
	var enums string

	{{range $element := .EnumVals}} 
	enums += "{{$element}}, "
	{{end}}
	enums = strings.TrimSuffix(enums, ", ")
	list := strings.Split(enums, ", ")
	flag := false

	for i := 0; i < len(list); i++ {
		if list[i] == params.{{.ParamName}} {
			flag = true
		}
	}

	if !flag  {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} must be one of [" + enums + "]\"}")
		return nil
	}
`))

	avStrDefaultTpl = template.Must(template.New("avStrDefaultTpl").Parse(`
	// paramDefaultStr_{{.StructType}}.{{.FieldName}}
	if params.{{.ParamName}} == "" {
		params.{{.ParamName}} = "{{.DefValStr}}"
	}
	`))

	avIntDefaultTpl = template.Must(template.New("avStrDefaultTpl").Parse(`
	// paramDefaultStr_{{.StructType}}.{{.FieldName}}
	if params.{{.ParamName}} == 0 {
		params.{{.ParamName}} = "{{.DefValInt}}"
	}
	`))

	avStrMaxCheckTpl = template.Must(template.New("avStrMaxCheckTpl").Parse(`
	// paramLenCheck_{{.StructType}}.{{.FieldName}}
	if len(params.{{.ParamName}}) > {{.MaxVal}} {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} len must be <= {{.MaxVal}}\"}")
		return nil
	}`))

	avIntMaxCheckTpl = template.Must(template.New("avStrMaxCheckTpl").Parse(`
	// paramLenCheck_{{.StructType}}.{{.FieldName}}
	if params.{{.ParamName}} > {{.MaxVal}} {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} must be <= {{.MaxVal}}\"}")
		return nil
	}`))

	avStrMinCheckTpl = template.Must(template.New("avStrMinCheckTpl").Parse(`
	if len(params.{{.ParamName}}) < {{.MinVal}} {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} len must be >= {{.MinVal}}\"}")
		return nil
	}
	`))

	avIntMinCheckTpl = template.Must(template.New("avStrMinCheckTpl").Parse(`
	if params.{{.ParamName}} < {{.MinVal}} {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"{{.FieldName}} must be >= {{.MinVal}}\"}")
		return nil
	}
	`))

	httpHandler = template.Must(template.New("httpHandler").Parse(`
// {{.FuncName}} http handler	
func (s *{{.ApiName}}) handler{{.FuncName}}(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if {{.IsPost}} {
		if isPost := postCheck(r); isPost != true {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Fprintln(w, "{\"error\":\"bad method\"}")
			return
		}
	}
	if {{.CheckAuth}} {
		if isAuth := checkAuth(r); !isAuth {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "{\"error\":\"unauthorized\"}")
			return
		}
	}
	//params := Val{{.FuncName}}Params{{.StructName}}(w, r)
	params := Val{{.StructName}}(w, r)

	if params != nil {
		resp, err := s.{{.FuncName}}(ctx, *params)
	
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
`))
)

func GenValidator(node *ast.File, out *os.File, StructName string, ApiName string) {
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.GenDecl\n", f)
			continue
		}
		for _, spec := range g.Specs {

			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
				continue
			}

			if currType.Name.Name != StructName {
				fmt.Printf("SKIP %T is not %v\n", currType.Name.Name, StructName)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}

			fmt.Printf("process struct %s\n", currType.Name.Name)
			fmt.Printf("\tgenerating val method\n")

			fmt.Fprintln(out, "func Val"+ currType.Name.Name +"(w http.ResponseWriter, r *http.Request) *" + currType.Name.Name  + " {")
			fmt.Fprintf(out, "\tvar params " + currType.Name.Name)

			for _, field := range currStruct.Fields.List {
				GenFieldChecks(node, out, field, currType.Name.Name)
			}
			fmt.Fprintln(out, "\n return &params")
			fmt.Fprintln(out, "} // End of validator for ", currType.Name.Name)
		}
	}
}

func GenFieldChecks(node *ast.File, out *os.File, field *ast.Field, curTypeName string) {
	result := "," + strings.TrimSuffix(strings.TrimPrefix(field.Tag.Value, "`apivalidator:\""), "\"`") + ","
	fieldType := field.Type.(*ast.Ident).Name

	pName := reParamname.FindString(result)
	if pName == "" {
		pName = strings.ToLower(field.Names[0].Name)
	} else {
		pName = strings.TrimPrefix(pName, ",paramname=")
	}
	required := reRequred.MatchString(result)

	pEnums := reEnum.FindString(result)

	pEnums = strings.TrimPrefix(pEnums, ",enum=")

	var list []string
	if pEnums != "" {
		list = strings.Split(pEnums, "|")
	}

	defVal := reDefault.FindString(result)
	defVal = strings.TrimPrefix(defVal, ",default=")

	min := reMin.FindString(result)
	min = strings.TrimPrefix(min, ",min=")

	max := reMax.FindString(result)
	max = strings.TrimPrefix(max,",max=")

	switch fieldType {
	case "int":
		intFillTpl.Execute(out, fillTpl{
			StructType: curTypeName,
			ParamName:  field.Names[0].Name,
			FieldName:  pName,
		})
		if required {
			avIntRequiredTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
			})
		}
		if defVal != "" {
			Val, _ := strconv.Atoi(defVal)
			err := avStrDefaultTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				DefValInt:  Val,
			})
			if err != nil {
				fmt.Println(err)
			}
		}
		if min != "" {
			min, err := strconv.Atoi(min)

			if err != nil {
				fmt.Println(err)
			}

			err = avIntMinCheckTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				EnumVals:   nil,
				DefValStr:  "",
				DefValInt:  0,
				MinVal:     min,
				MaxVal:     0,
			})
		}
		if max != "" {
			max, err := strconv.Atoi(max)

			if err != nil {
				fmt.Println(err)
			}

			err = avIntMaxCheckTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				EnumVals:   nil,
				DefValStr:  "",
				DefValInt:  0,
				MinVal:     0,
				MaxVal:     max,
			})
		}
	case "string":
		strFillTpl.Execute(out, fillTpl{
			StructType: curTypeName,
			ParamName:  field.Names[0].Name,
			FieldName:  pName,
		})
		if required {
			avStrRequiredTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
			})
		}
		if defVal != "" {
			err := avStrDefaultTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				DefValStr:  defVal,
			})
			if err != nil {
				fmt.Println(err)
			}
		}
		if len(pEnums) > 0 {
			err := avStrEnumTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				EnumVals:   list,
			})
			if err != nil {
				fmt.Println(err)
			}
		}
		if min != "" {
			min, err := strconv.Atoi(min)

			if err != nil {
				fmt.Println(err)
			}

			err = avStrMinCheckTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				EnumVals:   nil,
				DefValStr:  "",
				DefValInt:  0,
				MinVal:     min,
				MaxVal:     0,
			})
		}

		if max != "" {
			max, err := strconv.Atoi(max)

			if err != nil {
				fmt.Println(err)
			}

			err = avStrMaxCheckTpl.Execute(out, fillTpl{
				StructType: curTypeName,
				ParamName:  field.Names[0].Name,
				FieldName:  pName,
				EnumVals:   nil,
				DefValStr:  "",
				DefValInt:  0,
				MinVal:     0,
				MaxVal:     max,
			})
		}

	default:
		log.Fatalln("unsupported", fieldType)
	}
}
func GenHTTPServer(node *ast.File, out *os.File)  {
	var gen []string

	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.Decl\n", f)
			continue
		}
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}

			if ok := strings.HasSuffix(currType.Name.Name, "Api"); ok {
				gen = append(gen, currType.Name.Name)
			}
		}
	}

	for i := 0; i < len(gen); i++ {
		fmt.Fprintln(out, "func (h *" + gen[i] +") ServeHTTP(w http.ResponseWriter, r *http.Request) {")
		fmt.Fprintln(out, "switch r.URL.Path {")
		re := regexp.MustCompile("{(.*)}")
		for _, f := range node.Decls {
			fun, ok := f.(*ast.FuncDecl)
			if !ok {
				fmt.Printf("SKIP %T is not *ast.FuncDecl\n", f)
				continue
			}
			if fun.Doc == nil {
				fmt.Printf("SKIP func %#v doesnt have comments\n", fun.Name)
				continue
			}
			//if fun.Recv.List[0].Type
			// get description json
			if parent := fun.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name; parent != gen[i] {
				fmt.Printf("SKIP func %#v doesnt match argument list\n", parent)
				continue
			}
			match := re.FindString(fun.Doc.List[0].Text)
			var des descr

			err := json.Unmarshal([]byte(match), &des)

			if err != nil {
				fmt.Println(err)
			}
			name := strings.Split(des.URL, "/")
			fmt.Fprintln(out, "\tcase \""+des.URL+"\":")
			fmt.Fprintln(out, "\t\th.handler"+strings.Title(name[len(name)-1])+"(w, r)")
		}
		fmt.Fprintln(out, "\tdefault:")
		fmt.Fprintln(out, "\t\tw.WriteHeader(http.StatusNotFound)")

		fmt.Fprintln(out, "\t\tfmt.Fprintf(w, \"{\\\"error\\\":\\\"unknown method\\\"}\")")
		fmt.Fprintln(out, "\t}\n}\n")
	}
}

func GenHandlers(node *ast.File, out *os.File) {
	re := regexp.MustCompile("{(.*)}")
	for _, f := range node.Decls {
		fun, ok := f.(*ast.FuncDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.FuncDecl\n", f)
			continue
		}
		if fun.Doc == nil {
			fmt.Printf("SKIP func %#v doesnt have comments\n", fun.Name)
			continue
		}
		//if fun.Recv.List[0].Type
		// get description json
		if ok := strings.Contains(fun.Doc.List[0].Text, "apigen:api"); !ok {
			fmt.Printf("SKIP func %#v doesnt match argument list\n", fun.Name.Name)
			continue
		}
		match := re.FindString(fun.Doc.List[0].Text)
		var des descr

		err := json.Unmarshal([]byte(match), &des)

		if err != nil {
			fmt.Println(err)
		}

		var flag string
		if des.Method == "POST" {
			flag = "true"
		} else {
			flag = "false"
		}

		err = httpHandler.Execute(out, httpTpl{
			//StructName: fun.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name,
			StructName: fun.Type.Params.List[1].Type.(*ast.Ident).Name,
			ApiName:	fun.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name,
			FuncName:   fun.Name.Name,
			IsPost:     flag,
			CheckAuth:  strconv.FormatBool(des.Auth),
		})

		//fun.Recv.List
		/*
		for _, param := range fun.Type.Params.List {
			fmt.Println(src[param.Type.Pos()-offset:param.Type.End()-offset])
		}
		s := fun.Type.Params.List[1]
		fmt.Println(s[1].Names[0])*/
		s := fun.Type.Params.List[1].Type.(*ast.Ident).Name
		GenValidator(node, out, s, fun.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name)

		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out, `import (`)
	fmt.Fprintln(out, "\"encoding/json\"")
	fmt.Fprintln(out, "\"strings\"")
	fmt.Fprintln(out, "\"strconv\"")
	fmt.Fprintln(out, "\"fmt\"")
	fmt.Fprintln(out, "\"net/http\"")
	fmt.Fprintln(out, `)`)
	fmt.Fprintln(out) // empty line

	fmt.Fprintln(out, postCheck)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, authCheck)
	fmt.Fprintln(out, "type FR map[string]interface{}")

	GenHTTPServer(node, out)

	GenHandlers(node, out)
}