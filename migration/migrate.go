package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
)

//go:embed versions/*.sql
var versions embed.FS

//go:embed seed.sql
var seedSQL string

// Applique, dans l'ordre, les fichiers de migration/versions qui ne l'ont pas encore ete.
// Les versions deja jouees sont memorisees dans la table schema_migrations.
func Run(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(255) PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return fmt.Errorf("creation table schema_migrations : %w", err)
	}

	entries, err := versions.ReadDir("versions")
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	applied := 0
	for _, name := range names {
		var exists int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", name).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		content, err := versions.ReadFile("versions/" + name)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("migration %s : %w", name, err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name); err != nil {
			return err
		}
		log.Printf("Migration appliquee : %s", name)
		applied++
	}

	log.Printf("Migrations - %d appliquee(s), %d au total", applied, len(names))
	return nil
}

// Insere les donnees de test si la base est vide.
func Seed(db *sql.DB) error {
	var users int
	if err := db.QueryRow("SELECT COUNT(*) FROM utilisateurs").Scan(&users); err != nil {
		return fmt.Errorf("lecture utilisateurs (lancer migrate avant seed ?) : %w", err)
	}
	if users > 0 {
		log.Printf("Seed ignore : la base contient deja %d utilisateur(s)", users)
		return nil
	}

	if _, err := db.Exec(seedSQL); err != nil {
		return err
	}
	log.Printf("Seed - donnees de test inserees")
	return nil
}
