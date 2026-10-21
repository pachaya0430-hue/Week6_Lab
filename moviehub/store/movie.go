package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type MovieDetail struct {
	Movie
	Detail *string `json:"detail"`
}

func (s *Store) GetMovie(ctx context.Context, id int64) (MovieDetail, error) {
	var m MovieDetail

	err := s.Pool.QueryRow(ctx,
		`SELECT id, title, title_th, original_title, original_language,
		        overview, overview_th,
		        CASE WHEN poster_path LIKE 'http%' THEN poster_path
		             ELSE 'https://image.tmdb.org/t/p/w500' || poster_path END,
		        backdrop_path,
		        release_date::text, vote_average::double precision, vote_count,
		        popularity::double precision, adult, video,
		        coalesce(overview_th, overview)
		 FROM movies
		 WHERE id = $1`, id,
	).Scan(
		&m.ID,
		&m.Title,
		&m.TitleTH,
		&m.OriginalTitle,
		&m.OriginalLanguage,
		&m.Overview,
		&m.OverviewTH,
		&m.PosterPath,
		&m.BackdropPath,
		&m.ReleaseDate,
		&m.VoteAverage,
		&m.VoteCount,
		&m.Popularity,
		&m.Adult,
		&m.Video,
		&m.Detail,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrNotFound
	}

	return m, err
}
