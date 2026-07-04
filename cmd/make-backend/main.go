package makebackend

import (
	"database/sql"
	"log"
	"make-backend/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	// setup connection pool
	db, err := sql.Open("postgres", "ops")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := database.NewStore(db)
	store = store
}
