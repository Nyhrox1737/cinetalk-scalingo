package main

import (
	"forum/app"
	"forum/config"
	"forum/migration"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		runCommand(os.Args[1])
		return
	}

	application := app.InitApp()
	defer application.Close()

	address := ":" + application.Port
	log.Printf("Serveur lance : http://localhost%s", address)

	serveErr := http.ListenAndServe(address, application.Router)
	if serveErr != nil {
		log.Fatalf("Erreur lancement serveur - %s", serveErr.Error())
	}
}

// Sous-commandes : "migrate" applique les migrations, "seed" insere les donnees de depart.
func runCommand(name string) {
	config.LoadEnv()
	db := config.OpenDB(true)
	defer db.Close()

	var err error
	switch name {
	case "migrate":
		err = migration.Run(db)
	case "seed":
		err = migration.Seed(db)
	default:
		log.Fatalf("Commande inconnue : %s (attendu : migrate ou seed)", name)
	}
	if err != nil {
		log.Fatalf("Erreur %s - %s", name, err.Error())
	}
}
