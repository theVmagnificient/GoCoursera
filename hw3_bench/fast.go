package main

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)



// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open("data/users.txt")//filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	lines := strings.Split(string(fileContents), "\n")

	//users := make([]map[string]interface{}, 0)
	users := make([]JSONData, 0)
	for _, line := range lines {
		//user := make(map[string]interface{})
		user := JSONData{}
		// fmt.Printf("%v %v\n", err, line)
		//err = json.Unmarshal([]byte(line), &user)
		err = user.UnmarshalJSON([]byte(line))
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	var notSeenBefore bool
	for i, user := range users {

		isAndroid := false
		isMSIE := false

		//browsers, ok := user["browsers"].([]interface{})
		browsers := user.Browsers
		/*if !ok {
			// log.Println("cant cast browsers")
			continue
		}*/

		for _, browserRaw := range browsers {
			//browser, ok := browserRaw.(string)
			browser := browserRaw
			/*if !ok {
				// log.Println("cant cast browser to string")
				continue
			}*/
			if ok := strings.Contains(browser, "Android" ); ok {//regexp.MatchString("Android", browser); ok && err == nil {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range browsers {
			//browser, ok := browserRaw.(string)
			browser := browserRaw
			/* if !ok {
				// log.Println("cant cast browser to string")
				continue
			} */
			if ok := strings.Contains(browser, "MSIE"); ok {//regexp.MatchString("MSIE", browser); ok && err == nil {
				isMSIE = true
				notSeenBefore = true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		//email := r.ReplaceAllString(user["email"].(string), " [at] ")
		//email := strings.Replace(user["email"].(string), "@", " [at] ", -1)
		email := strings.Replace(user.Email, "@", " [at] ", -1)
		//foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

