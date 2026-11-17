# CinéTalk 🎬

CinéTalk est une plateforme communautaire de discussion dédiée au cinéma et aux séries,
inspirée de services comme Stack Overflow, Reddit et IMDb. Les membres peuvent ouvrir des
fils de discussion, échanger des messages, réagir aux contributions de la communauté et
organiser les contenus par genres.

**Application en ligne : https://cinetalk-nohan.osc-fr1.scalingo.io**

Route de santé : https://cinetalk-nohan.osc-fr1.scalingo.io/health

## Membres de l'équipe

- Shakil Khaldi
- Hugo Berton
- Nohan Reis

## Fonctionnalités

- Inscription et connexion (par nom d'utilisateur ou e-mail) avec mots de passe hachés en SHA-512 + sel
- Authentification par jeton **JWT** stocké dans un cookie sécurisé (HttpOnly, Secure en HTTPS)
- Création, modification et suppression de fils de discussion
- Gestion des états d'un fil : `ouvert`, `fermé`, `archivé`
- Publication, modification et suppression de messages
- Système de réactions like / dislike avec calcul d'un score de popularité (AJAX)
- Tri des messages : plus récents, chronologique ou par popularité
- Pagination configurable (10, 20, 30 ou tous les éléments)
- Association des fils à un ou plusieurs genres / catégories et filtrage par genre
- Recherche de fils par titre ou par genre (réservée aux membres connectés)
- **Catalogue TMDB** : films, séries, acteurs et réalisateurs via l'API The Movie Database
- Tableau de bord d'administration : modération des fils, suppression de contenus,
  changement d'état et bannissement de comptes

## Technologies utilisées

- **Go 1.25** (routes, contrôleurs, logique métier, accès aux données, sécurité)
- **gorilla/mux** pour le routage, **html/template** pour les vues
- **MySQL 8** (base de données relationnelle, addon managé sur Scalingo)
- **golang-jwt** pour les jetons JWT
- **Scalingo** pour l'hébergement, **GitHub Actions** pour l'intégration continue

## Architecture de production

```
                 HTTPS uniquement
 Navigateur  ─────────────────────▶  App Scalingo "cinetalk-nohan"
                                     buildpack Go, 1 conteneur S
                                     binaire `forum` sur $PORT
                                     GET /health
                                              │
                                              │ SQL (SCALINGO_MYSQL_URL)
                                              ▼
                                     Addon MySQL Starter 512
                                     sauvegardes périodiques
```

### Pourquoi une seule application

Le sujet prévoit deux applications (client React + API). CinéTalk est un monolithe Go qui
rend ses pages HTML côté serveur : il n'y a pas de client séparé à builder, et le
JavaScript (réactions) appelle l'API sur la même origine. Une seule app Scalingo suffit
donc, ce qui correspond à la variante acceptée par le sujet. Conséquences :

- pas de configuration CORS (même origine) ;
- un seul conteneur à payer et à surveiller ;
- le conteneur reste éphémère : aucune donnée n'est écrite sur disque, tout est en base.

### Fichiers liés au déploiement

| Fichier | Rôle |
|---------|------|
| `Procfile` | process `web` (le serveur) et `postdeploy` (migrations après chaque déploiement) |
| `migration/versions/*.sql` | migrations de schéma versionnées, jouées une seule fois chacune |
| `migration/migrate.go` | exécuteur de migrations (table `schema_migrations`) et seed |
| `migration/seed.sql` | données de départ, insérées une seule fois via `scalingo run forum seed` |
| `config/db.config.go` | lit `SCALINGO_MYSQL_URL` en production, les variables `DB_*` en local |
| `.github/workflows/ci.yml` | `go vet`, `go test`, `go build` à chaque push et pull request |
| `docker-compose.yml` | MySQL en local, identique à la production |

## Variables d'environnement

Aucune valeur secrète n'est dans Git. Le fichier `.env.example` liste les variables ;
en local, copier-le en `.env`. Sur Scalingo, elles sont définies dans l'onglet
*Environment* ou avec `scalingo env-set`.

| Variable | Description | Production |
|----------|-------------|------------|
| `PORT` | port d'écoute du serveur | fourni automatiquement par Scalingo |
| `SCALINGO_MYSQL_URL` | URL complète de la base (`mysql://user:mdp@hote:port/base`) | fournie par l'addon MySQL |
| `DB_NAME`, `DB_USER`, `DB_PWD`, `DB_HOST`, `DB_PORT` | connexion MySQL en local (ignorées si `SCALINGO_MYSQL_URL` existe) | non utilisées |
| `JWT_SECRET` | clé de signature des jetons JWT | à définir (chaîne aléatoire longue) |
| `TMDB_API_KEY` | clé API The Movie Database pour le catalogue | à définir (facultative, le forum fonctionne sans) |

## Lancer le projet en local

Prérequis : [Go 1.25](https://go.dev/dl/) et [Docker](https://www.docker.com/).

```bash
git clone https://github.com/Nyhrox1737/cinetalk-scalingo.git
cd cinetalk-scalingo
cp .env.example .env          # puis renseigner DB_PWD=cinetalk et un JWT_SECRET

docker compose up -d db       # MySQL 8 dans un conteneur
go run . migrate              # crée les tables
go run . seed                 # insère les données de test (une seule fois)
go run .                      # http://localhost:8080
```

Alternative tout-en-un : `docker compose up --build` lance la base et l'application
(les migrations sont jouées au démarrage du conteneur, le seed reste manuel :
`docker compose exec app ./forum seed`).

## Procédure de déploiement (reproductible)

1. Installer la [CLI Scalingo](https://doc.scalingo.com/platform/cli/start) puis se connecter
   et ajouter sa clé SSH :
   ```bash
   scalingo login
   scalingo keys-add mon-pc ~/.ssh/id_ed25519.pub
   ```
2. Créer l'application et sa base (région Paris) :
   ```bash
   scalingo create cinetalk-nohan --region osc-fr1
   scalingo --app cinetalk-nohan addons-add mysql mysql-starter-512
   ```
3. Définir les secrets et forcer le HTTPS :
   ```bash
   scalingo --app cinetalk-nohan env-set JWT_SECRET=$(openssl rand -hex 32)
   scalingo --app cinetalk-nohan env-set TMDB_API_KEY=<votre_cle>
   scalingo --app cinetalk-nohan force-https --enable
   ```
4. Déployer : le build compile le binaire Go, puis le process `postdeploy` applique les
   migrations.
   ```bash
   git push scalingo main
   ```
5. Dimensionner au plus juste et insérer les données de départ (une seule fois) :
   ```bash
   scalingo --app cinetalk-nohan scale web:1:S
   scalingo --app cinetalk-nohan run --detached forum seed
   ```
6. Vérifier :
   ```bash
   curl https://cinetalk-nohan.osc-fr1.scalingo.io/health
   # {"database":"ok","status":"ok"}
   ```

### Déploiement continu

L'application est liée au dépôt GitHub `Nyhrox1737/cinetalk-scalingo` avec *Auto Deploy*
sur la branche `main`. À chaque push, la GitHub Action `CI` lance `go vet`, `go test` et
`go build` ; Scalingo attend que l'Action soit verte avant de déployer, un code rouge
n'atteint donc jamais la production.

Mise en place : dans le dashboard Scalingo, onglet *Deploy* → *GitHub* → autoriser
le compte GitHub, choisir le dépôt, cocher *Automatic deploys* sur `main`.
Équivalent CLI une fois le compte lié (`scalingo integrations-add github`) :

```bash
scalingo --app cinetalk-nohan integration-link-create --auto-deploy --branch main https://github.com/Nyhrox1737/cinetalk-scalingo
```

## Coût d'hébergement

Estimation faite avec le simulateur de prix Scalingo avant la création des ressources
(prix HT, région osc-fr1) :

| Ressource | Plan | Prix mensuel |
|-----------|------|--------------|
| Conteneur web | 1 × S (256 Mo) | 7,20 € |
| Base de données | MySQL Starter 512 (512 Mo, 10 Go) | 7,20 € |
| **Total estimé** | | **14,40 € HT / mois** |

La facturation est à la minute : l'application tourne depuis le 17 novembre 2026.
Consommation réelle relevée dans *Billing* : à compléter chaque semaine
(17/11/2026 : 0,00 €, ressources créées le jour même).

Le voucher Ynov couvre 50 € par mois ; ce dimensionnement laisse une marge confortable.
En fin de module, mettre le conteneur à 0 (`scalingo --app cinetalk-nohan scale web:0`)
ou supprimer l'application pour arrêter la facturation.

## En cas de panne

Toutes les commandes ci-dessous ont été testées sur l'application.

### Lire les logs

```bash
scalingo --app cinetalk-nohan logs --lines 200          # logs de l'application
scalingo --app cinetalk-nohan logs -f                   # en continu
scalingo --app cinetalk-nohan logs --filter postdeploy  # uniquement les migrations
scalingo --app cinetalk-nohan logs --addon ad-6778280c-ffdf-4b9d-9d2b-8f34ed2d8911 --lines 100   # logs MySQL
scalingo --app cinetalk-nohan ps                        # état des conteneurs
scalingo --app cinetalk-nohan stats                     # CPU et mémoire en direct
```

Exemple de diagnostic : lancer `scalingo --app cinetalk-nohan run --detached forum boom`
démarre un conteneur one-off avec une commande inconnue. Les logs affichent
`Commande inconnue : boom (attendu : migrate ou seed)` avec le nom du conteneur
(`one-off-XXXX`), ce qui permet de retrouver la cause en filtrant :
`scalingo --app cinetalk-nohan logs --filter one-off-XXXX`.

Autre cas typique : si `/health` renvoie `503 {"database":"ko"}`, la base est
injoignable ; le détail est dans les logs de l'application, et l'état de l'addon se voit
avec `scalingo --app cinetalk-nohan addons`.

### Revenir à un déploiement précédent (rollback)

Scalingo redéploie n'importe quel commit poussé sur son remote. Pour revenir en arrière :

```bash
scalingo --app cinetalk-nohan deployments        # historique avec les SHA git
git push --force scalingo <sha_qui_marchait>:main
```

Le build se relance sur l'ancien commit et remplace le conteneur en cours uniquement
s'il réussit. Pour revenir ensuite à la dernière version : `git push --force scalingo main`.

### Sauvegarder la base

Les sauvegardes périodiques sont activées sur l'addon. Pour en faire une à la demande
et la télécharger (archive `.tar.gz` contenant un dump MySQL) :

```bash
scalingo --app cinetalk-nohan --addon ad-6778280c-ffdf-4b9d-9d2b-8f34ed2d8911 backups
scalingo --app cinetalk-nohan --addon ad-6778280c-ffdf-4b9d-9d2b-8f34ed2d8911 backups-create
scalingo --app cinetalk-nohan --addon ad-6778280c-ffdf-4b9d-9d2b-8f34ed2d8911 backups-download
```

### Restaurer la base

1. Ouvrir un tunnel chiffré vers la base (laisser la commande tourner) :
   ```bash
   scalingo --app cinetalk-nohan db-tunnel SCALINGO_MYSQL_URL
   ```
   La base est alors accessible sur `127.0.0.1:10000` avec l'utilisateur et le mot de
   passe de `SCALINGO_MYSQL_URL` (`scalingo --app cinetalk-nohan env | grep MYSQL`).
2. Dans un autre terminal, réinjecter le dump de la sauvegarde :
   ```bash
   tar -xzf <sauvegarde>.tar.gz
   mysql -h 127.0.0.1 -P 10000 -u <user> -p <base> < <fichier>.sql
   ```
3. Vérifier avec `curl https://cinetalk-nohan.osc-fr1.scalingo.io/health` puis en ouvrant le site.

Pour repartir d'une base vide (sans sauvegarde) : `scalingo --app cinetalk-nohan run forum migrate`
puis `scalingo --app cinetalk-nohan run forum seed`.

## Comptes de test

| Identifiant | Mot de passe      | Rôle           |
|-------------|-------------------|----------------|
| `admin`     | `Admin@CineTalk1` | Administrateur |
| `alice`     | `Cinephile@2026`  | Utilisateur    |
| `bob`       | `Cinephile@2026`  | Utilisateur    |
| `claire`    | `Cinephile@2026`  | Utilisateur    |
| `david`     | `Cinephile@2026`  | Utilisateur    |

## Documentation et gestion de projet

- **Diagrammes** (dossier [`docs/`](docs/)) : architecture C4, cas d'utilisation et modèle
  logique de données (MLD).
- **Suivi des tâches** : [tableau Trello du projet](https://trello.com/b/rzDcTCB9/mon-tableau-trello).

## Sécurité

- Les mots de passe ne sont jamais stockés en clair : hachage **SHA-512** avec un sel unique
  par utilisateur.
- Mot de passe : 12 caractères minimum, une majuscule et un caractère spécial.
- Le jeton JWT est vérifié à chaque requête ; les comptes bannis perdent immédiatement l'accès.
- Aucun secret dans le dépôt : tout passe par les variables d'environnement.
