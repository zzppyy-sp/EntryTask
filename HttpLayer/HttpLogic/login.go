package main

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"google.golang.org/protobuf/proto"
	pool "pengyu.zhao/EntryTask/Http/HttpLogic/connectPool"
)

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

var connectPool pool.Pool

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type profilePageData struct {
	UserName   string
	NickName   string
	PicAddress string
}

func initConnectionPool() pool.Pool {
	factory := func() (net.Conn, error) { return net.Dial("tcp", "localhost:8080") }
	p, err := pool.NewChannelPool(5, 30, factory)
	if err != nil {
		panic(err)
	}
	return p
}
func intToByteArray(a uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, a)
	return bs
}

func verifyUser(userID string, password string) (bool, error) {
	conn, err := connectPool.Get()

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	comm := &Command{
		CommandType: Command_VERIFYUSER,
		User: &User{
			ID:       userID,
			Password: password,
		},
	}

	data, err := proto.Marshal(comm)
	if err != nil {
		panic(err)
	}
	fmt.Println("size of data sending is :", len(data))
	dataLen := intToByteArray(uint32(len(data)))
	conn.Write(dataLen)
	conn.Write(data)

	//read response from TCP server
	readPackages(conn)

	return true, nil

}

func readPackages(c net.Conn) *Response {
	res := &Response{}
	dataLenBytes := make([]byte, 4)
	_, err1 := io.ReadFull(c, dataLenBytes)
	if err1 != nil {
		panic(err1)
	}
	resLen := binary.LittleEndian.Uint32(dataLenBytes)
	fmt.Println("size of data reading is:", resLen)
	resBytes := make([]byte, resLen)

	_, err2 := io.ReadFull(c, resBytes)
	if err2 != nil {
		panic(err1)
	}

	err3 := proto.Unmarshal(resBytes, res)
	if err3 != nil {
		panic(err1)
	}
	fmt.Println("response type is :", res.GetResponseType())
	return res
}

func signin(w http.ResponseWriter, r *http.Request) {
	// var creds Credentials
	// Get the JSON body and decode into credentials
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../templates/login.html")
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		r.ParseForm()
		inputUserID := r.FormValue("userID")
		inputPassword := r.FormValue("password")
		isValid, err := verifyUser(inputUserID, inputPassword)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		if !isValid || err != nil {
			w.Write([]byte("error"))
		}

	}

	// err := json.NewDecoder(r.Body).Decode(&creds)
	// if err != nil {
	// 	// If the structure of the body is wrong, return an HTTP error
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// // Get the expected password from our in memory map
	// expectedPassword, ok := users[creds.Username]

	// // If a password exists for the given user
	// // AND, if it is the same as the password we received, the we can move ahead
	// // if NOT, then we return an "Unauthorized" status
	// if !ok || expectedPassword != creds.Password {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	return
	// }

	// Create a new random session token
	// sessionToken := uuid.New().String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds

	// _, err = cache.Do("SETEX", sessionToken, "120", creds.Username)
	// if err != nil {
	// 	// If there is an error in setting the cache, return an internal server error
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds, the same as the cache
	// http.SetCookie(w, &http.Cookie{
	// 	Name:    "session_token",
	// 	Value:   sessionToken,
	// 	Expires: time.Now().Add(120 * time.Second),
	// })
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	// fmt.Println(r.Form) // print information on server side.
	// fmt.Println("path", r.URL.Path)
	// fmt.Println("scheme", r.URL.Scheme)
	// fmt.Println(r.Form["url_long"])
	// for k, v := range r.Form {
	// 	fmt.Println("key:", k)
	// 	fmt.Println("val:", strings.Join(v, ""))
	// }
	fmt.Fprintf(w, "Hello astaxie!") // write data to response
}

func showUserProfile(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, _ := template.ParseFiles("../templates/userPage.html")
		data := profilePageData{
			UserName:   "hello",
			NickName:   "world",
			PicAddress: "/documents/image.jpg",
		}
		t.Execute(w, data)
	} else if r.Method == "POST" {
		fmt.Println("File Upload Endpoint Hit")

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		r.ParseMultipartForm(10 << 20)
		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)

		// Create a temporary file within our temp-images directory that follows
		// a particular naming pattern
		tempFile, err := ioutil.TempFile("../documents", "upload-*.png")
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)
		// return that we have successfully uploaded our file!
		fmt.Fprintf(w, "Successfully Uploaded File\n")
	}
}

func main() {
	connectPool = initConnectionPool()

	http.HandleFunc("/", sayhelloName) // setting router rule
	http.HandleFunc("/login", signin)
	http.HandleFunc("/userProfile", showUserProfile)
	http.Handle("/documents/", http.StripPrefix("/documents/", http.FileServer(http.Dir("../documents"))))

	err := http.ListenAndServe(":9090", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
