package quotes

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func GetRandomQuote(ctx context.Context) (*Quote, error) {
	quote, err := db.Get[Quote](ctx, "SELECT id, text, author FROM quotes ORDER BY RANDOM() LIMIT 1")
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func GetAllQuotes(ctx context.Context) ([]Quote, error) {
	return db.Select[Quote](ctx, "SELECT id, text, author FROM quotes ORDER BY id")
}

func GetQuoteByID(ctx context.Context, id int) (*Quote, error) {
	quote, err := db.Get[Quote](ctx, "SELECT id, text, author FROM quotes WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func CreateQuote(ctx context.Context, text, author string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO quotes (text, author)
		VALUES ($1, $2)
		RETURNING id
	`, text, author).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create quote: %w", err)
	}
	return id, nil
}

func UpdateQuote(ctx context.Context, id int, text, author string) error {
	_, err := db.Exec(ctx, `
		UPDATE quotes
		SET text = $1, author = $2
		WHERE id = $3
	`, text, author, id)
	if err != nil {
		return fmt.Errorf("failed to update quote: %w", err)
	}
	return nil
}

func DeleteQuote(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, "DELETE FROM quotes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete quote: %w", err)
	}
	return nil
}
