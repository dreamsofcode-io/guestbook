package guest

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{
		db: db,
	}
}

var insertSQL = `
INSERT INTO guest (id, message, created_at, updated_at, ip)
VALUES ($1, $2, $3, $3, $4)
`

func (r *Repo) Insert(ctx context.Context, guest Guest) error {
	_, err := r.db.Exec(
		ctx, insertSQL, guest.ID, guest.Message, guest.CreatedAt.UTC(),
		guest.IP,
	)
	if err != nil {
		return fmt.Errorf("execute sql: %w", err)
	}

	return nil
}

var selectSQL = `
SELECT id, message, created_at, ip
FROM guest
ORDER BY created_at DESC
LIMIT $1
`

func (r *Repo) FindAll(ctx context.Context, count int) ([]Guest, error) {
	rows, err := r.db.Query(ctx, selectSQL, count)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	res := []Guest{}

	for rows.Next() {
		var guest Guest

		err = rows.Scan(&guest.ID, &guest.Message, &guest.CreatedAt, &guest.IP)
		if err != nil {
			fmt.Println(err)
			continue
		}

		guest.CreatedAt = guest.CreatedAt.UTC()

		res = append(res, guest)
	}

	return res, nil
}

var countSQL = `
SELECT COUNT(*) FROM guest
`

func (r *Repo) Count(ctx context.Context) (int, error) {
	count := 0

	err := r.db.QueryRow(ctx, countSQL).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query row: %w", err)
	}

	return count, nil
}
