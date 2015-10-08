package main

import (
	"fmt"
	"net/mail"

	"github.com/coduno/api/model"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
)

const projID = "coduno"

var ctx = context.Background()

func main() {
	conf, err := google.NewSDKConfig("")
	if err != nil {
		panic(err)
	}

	fmt.Println("Configuration loaded.")

	client, err := datastore.NewClient(ctx, projID, cloud.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		panic(err)
	}

	fmt.Println("Client created.")

	q := datastore.NewQuery(model.CompanyKind).Filter("Name =", "Frequentis")

	var company model.Company

	companyKey, err := client.Run(ctx, q).Next(&company)
	if err != nil {
		panic(err)
	}

	fmt.Println("Company found.")

	hpw, err := bcrypt.GenerateFromPassword([]byte("S7gxqLeGvgv99mDnLp7P"), 0)
	if err != nil {
		panic(err)
	}

	fmt.Println("Password hashed.")

	user := model.User{
		Address: mail.Address{
			Name:    "Mag. Anne-Kathrin Lindner",
			Address: "anne-kathrin.lindner@frequentis.com",
		},
		Nick:           "",
		Company:        companyKey,
		HashedPassword: hpw,
	}

	_, err = client.Put(ctx, datastore.NewIncompleteKey(ctx, model.UserKind, nil), &user)
	if err != nil {
		panic(err)
	}

	fmt.Println("User injected.")
}
