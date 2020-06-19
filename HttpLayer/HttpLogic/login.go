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

	"github.com/gorilla/sessions"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/proto"
	pool "pengyu.zhao/EntryTask/Http/HttpLogic/connectPool"
)

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

var connectPool pool.Pool
var store = sessions.NewCookieStore([]byte("t0p-s3cr3t"))

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

func getProfilePageDataFromSession(seesion *sessions.Session, data *profilePageData) {
	untyped, ok := seesion.Values["userName"]
	if !ok {
		return
	}
	userName, ok := untyped.(string)
	if !ok {
		return
	}
	untyped, ok = seesion.Values["nickName"]
	if !ok {
		return
	}
	nickName, ok := untyped.(string)
	if !ok {
		return
	}
	untyped, ok = seesion.Values["pictureAddress"]
	if !ok {
		return
	}
	pictureAddress, ok := untyped.(string)
	if !ok {
		return
	}
	data.NickName = nickName
	data.UserName = userName
	data.PicAddress = "/documents/" + pictureAddress

}

func verifyUser(userID string, password string) *Response {
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
	res := readPackages(conn)

	return res

}

func retrieveUerInfo(userID string) *Response {
	conn, err := connectPool.Get()
	if err != nil {
		panic(err)
	}
	comm := &Command{
		CommandType: Command_RETREIVEUSERINFO,
		User: &User{
			ID: userID,
		},
	}

	data, err := proto.Marshal(comm)
	if err != nil {
		panic(err)
	}
	fmt.Println("sending retrieve userInfo package....")
	fmt.Println("size of data sending is :", len(data))
	dataLen := intToByteArray(uint32(len(data)))
	conn.Write(dataLen)
	conn.Write(data)

	//read response from TCP server
	res := readPackages(conn)

	return res

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
	fmt.Println("welcome to sign in page")
	var isLogged bool = true
	cookie, err := r.Cookie("session")
	if err != nil {
		isLogged = false
		id := uuid.NewV4()
		cookie = &http.Cookie{
			Name:  "session",
			Value: id.String(),
		}
		http.SetCookie(w, cookie)
	}

	if r.Method == "GET" {
		if isLogged {
			http.Redirect(w, r, "/userProfile", http.StatusSeeOther)
		} else {
			t, _ := template.ParseFiles("../templates/login.html")
			t.Execute(w, nil)
		}
	} else if r.Method == "POST" {
		r.ParseForm()
		inputUserID := r.FormValue("userID")
		inputPassword := r.FormValue("password")
		isValid := verifyUser(inputUserID, inputPassword)
		if isValid.GetResponseType() != Response_SUCCESSFULL {
			io.WriteString(w, "invalid user or password")
		} else {
			res := retrieveUerInfo(inputUserID)

			session, _ := store.Get(r, cookie.Value)
			session.Values["userID"] = inputUserID
			session.Values["userName"] = res.GetUser().GetUserName()
			session.Values["nickName"] = res.GetUser().GetNickName()
			session.Values["pictureAddress"] = res.GetUser().GetPictureAddress()
			session.Save(r, w)
			http.Redirect(w, r, "/userProfile", http.StatusSeeOther)
		}

	}

}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func showUserProfile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		fmt.Println("cookie is empty")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	sID := cookie.Value
	session, err := store.Get(r, sID)
	if err != nil {
		panic(err)
	}
	data := &profilePageData{}
	getProfilePageDataFromSession(session, data)
	fmt.Println(r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../templates/userPage.html")
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
