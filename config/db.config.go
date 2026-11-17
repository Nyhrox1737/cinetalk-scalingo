package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Construit la chaine de connexion MySQL.
// Sur Scalingo, l'addon fournit SCALINGO_MYSQL_URL (mysql://user:mdp@hote:port/base).
// En local, on utilise les variables DB_USER, DB_PWD, DB_HOST, DB_PORT et DB_NAME.
func BuildDSN(multiStatements bool) string {
	params := "charset=utf8mb4&collation=utf8mb4_unicode_ci"
	if multiStatements {
		params += "&multiStatements=true"
	}

	if raw := GetEnvWithDefault("SCALINGO_MYSQL_URL", GetEnvWithDefault("DATABASE_URL", "")); raw != "" {
		parsed, err := url.Parse(raw)
		if err != nil {
			log.Fatalf("Erreur configuration - URL de base de donnees invalide : %s", err.Error())
		}
		pwd, _ := parsed.User.Password()
		name := strings.TrimPrefix(parsed.Path, "/")
		return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", parsed.User.Username(), pwd, parsed.Host, name, params)
	}

	user := GetRequiredEnv("DB_USER")
	pwd := GetEnvWithDefault("DB_PWD", "")
	host := GetRequiredEnv("DB_HOST")
	port := GetRequiredEnv("DB_PORT")
	name := GetRequiredEnv("DB_NAME")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, pwd, host, port, name, params)
}

// Ouvre la connexion a la base et verifie qu'elle repond.
func OpenDB(multiStatements bool) *sql.DB {
	dbContext, dbContextErr := sql.Open("mysql", BuildDSN(multiStatements))
	if dbContextErr != nil {
		log.Fatalf("Erreur connexion base de donnees - Erreur : \n\t %s", dbContextErr.Error())
	}

	pingErr := dbContext.Ping()
	if pingErr != nil {
		dbContext.Close()
		log.Fatalf("Erreur ping base de donnees - Erreur : \n\t %s", pingErr.Error())
	}

	log.Printf("BDD - Connexion reussie")
	return dbContext
}

// Connexion utilisee par l'application web.
func InitDB() *sql.DB {
	return OpenDB(false)
}
