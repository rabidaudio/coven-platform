// Admin CLI exposes operations for managing the backend.
//
// Usage:
//
//	admin [command] [flags]
//
// The commands are:
//
//	register
//			Create a new user with a random password. The user will need
//			to use the password reset function to change it.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/atlantacoven/coven-platform/member-site/database"
	"github.com/atlantacoven/coven-platform/member-site/users"
)

var register = flag.NewFlagSet("register", flag.ExitOnError)
var registerName = register.String("name", "", "the name of the user")
var registerEmail = register.String("email", "", "the email address of the user (required)")
var registerPassword = register.String("pass", "", "the password for the user. if not set, generates a random one")

var cmds = []*flag.FlagSet{
	register,
}

func main() {
	db := must(database.Create())
	defer db.Close()

	ctx := database.WithDB(db, context.Background())

	if len(os.Args) < 2 {
		fmt.Println("command required")
		ShowHelp()
	}

	cmd := os.Args[1]
	switch cmd {
	case "help":
		ShowHelp()
	case "register":
		register.Parse(os.Args[2:])
		Register(ctx)

	default:
		fmt.Println("unknown command")
		ShowHelp()
	}
}

func Register(ctx context.Context) {
	if *registerEmail == "" {
		fmt.Println("email is required")
		ShowHelp()
	}
	pass := *registerPassword
	if pass == "" {
		pass = randomPassword()
	}

	u, err := users.Register(ctx, *registerName, *registerEmail, pass)
	if err != nil {
		fmt.Printf("registration failed: %v\n", err)
		os.Exit(10)
	}
	fmt.Printf("User %v registered.\n", u.Id)
	if *registerPassword == "" {
		fmt.Printf("Temporary password: %v\n", pass)
	}
}

func ShowHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of admin:\n")
	for _, cmd := range cmds {
		fmt.Printf("\nadmin %v\n", cmd.Name())
		cmd.PrintDefaults()
	}
	flag.PrintDefaults()
	os.Exit(1)
}

func randomPassword() string {
	passbytes := make([]byte, 8)
	rand.Read(passbytes)
	return base64.StdEncoding.EncodeToString(passbytes)
}

func must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}
