package dbConnect

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func BuildConnection() *sql.DB {
	db, err := sql.Open("mysql", "root:zpy19980412@tcp(localhost:3306)/EntryTask")
	if err != nil {
		panic(err.Error())
	}
	return db
}

func RetriveUserData(db *sql.DB, userID string) (string, string, string) {

	query := "select userName,nickName,profilePicAddress from users where id = " + userID + ""
	userName := ""
	nickName := ""
	profilePicAddress := ""
	err := db.QueryRow(query).Scan(&userName, &nickName, &profilePicAddress)
	if err != nil {
		panic(err)
	}
	return userName, nickName, profilePicAddress
}

func UpdateNickName(db *sql.DB, userID string, nickName string) error {
	query := "update users set nickName = '" + nickName + "' where id = '" + userID + "'"
	// fmt.Println(query)

	update, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer update.Close()
	return nil
}

func UpdatePic(db *sql.DB, userID string, picAddress string) error {
	query := "update users set  profilePicAddress = '" + picAddress + "' where id = '" + userID + "'"
	// fmt.Println(query)

	update, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer update.Close()
	return nil
}

func VerifyIdentity(db *sql.DB, userID string, pass string) (bool, error) {
	password := ""
	query := "select pass from users where id = '" + userID + "'"
	// fmt.Println(query)

	sel, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&password)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(tag.ID, tag.Pass)
	}

	if password == "" {
		return false, fmt.Errorf("User does not exist")
	} else if password != pass {
		return false, fmt.Errorf("Incorect password")
	} else {
		return true, nil
	}

}
