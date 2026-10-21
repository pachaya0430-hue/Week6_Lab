package store

import (
	"context"
	"time"
)

type Movie struct {
	ID               int64   `json:"id"`
	Title            string  `json:"title"`
	TitleTH          *string `json:"title_th,omitempty"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Overview         *string `json:"overview,omitempty"`
	OverviewTH       *string `json:"overview_th,omitempty"`
	PosterPath       *string `json:"poster_path,omitempty"`
	BackdropPath     *string `json:"backdrop_path,omitempty"`
	ReleaseDate      *string `json:"release_date,omitempty"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	Adult            bool    `json:"adult"`
	Video            bool    `json:"video"`
}

type Review struct {
	ID          int64     `json:"id"`
	DisplayName string    `json:"displayName"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (s *Store) SearchMovies(ctx context.Context, query string) ([]Movie, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, title, title_th, original_title, original_language,
		        overview, overview_th,
		        CASE WHEN poster_path LIKE 'http%' THEN poster_path
		             ELSE 'https://image.tmdb.org/t/p/w500' || poster_path END,
		        backdrop_path,
		        release_date::text, vote_average::double precision, vote_count,
		        popularity::double precision, adult, video
		 FROM movies
		 WHERE $1 = '' OR title ILIKE '%' || $1 || '%'
		    OR title_th ILIKE '%' || $1 || '%'
		    OR original_title ILIKE '%' || $1 || '%'
		 ORDER BY popularity DESC, id
		 LIMIT 50`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := make([]Movie, 0)
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(
			&movie.ID, &movie.Title, &movie.TitleTH, &movie.OriginalTitle,
			&movie.OriginalLanguage, &movie.Overview, &movie.OverviewTH,
			&movie.PosterPath, &movie.BackdropPath, &movie.ReleaseDate,
			&movie.VoteAverage, &movie.VoteCount, &movie.Popularity,
			&movie.Adult, &movie.Video,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (s *Store) ListReviews(ctx context.Context, movieID int64) ([]Review, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT r.id, m.display_name, r.text, r.created_at, r.updated_at
		 FROM reviews AS r
		 JOIN members AS m ON m.id = r.member_id
		 WHERE r.movie_id = $1
		 ORDER BY r.created_at DESC, r.id DESC`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)
	for rows.Next() {
		var review Review
		if err := rows.Scan(
			&review.ID, &review.DisplayName, &review.Text,
			&review.CreatedAt, &review.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reviews, nil
}
