package main

import "time"

type Poem struct {
	Slug      string    `db:"slug"`
	Title     string    `db:"title"`
	Poem      string    `db:"poem"`
	Notes     string    `db:"notes"`
	Published time.Time `db:"published"`
	Modified  time.Time `db:"modified"`
}
