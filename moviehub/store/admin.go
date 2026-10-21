package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicateID = errors.New("duplicate id")
	ErrBadGenre    = errors.New("bad genre")
)

const posterSQL = `CASE WHEN m.poster_path LIKE 'http%' THEN m.poster_path
                        ELSE 'https://image.tmdb.org/t/p/w500' || m.poster_path END`

type Genre struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	NameTH *string `json:"nameTh,omitempty"`
}

type TopMovie struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	TitleTh     *string `json:"titleTh"`
	Poster      *string `json:"poster"`
	VoteCount   int64   `json:"voteCount"`
	VoteAverage float64 `json:"voteAverage"`
}

type Stats struct {
	Movies        int64      `json:"movies"`
	Members       int64      `json:"members"`
	Admins        int64      `json:"admins"`
	Votes         int64      `json:"votes"`
	WishlistItems int64      `json:"wishlistItems"`
	NewestMembers []Member   `json:"newestMembers"`
	TopVoted      []TopMovie `json:"topVoted"`
}

func (s *Store) Stats(ctx context.Context) (Stats, error) {
	var stats Stats
	err := s.Pool.QueryRow(ctx,
		`SELECT (SELECT count(*) FROM movies),
		        (SELECT count(*) FROM members),
		        (SELECT count(*) FROM members WHERE role = 'admin'),
		        (SELECT count(*) FROM votes),
		        (SELECT count(*) FROM wishlist_items)`).
		Scan(&stats.Movies, &stats.Members, &stats.Admins, &stats.Votes, &stats.WishlistItems)
	if err != nil {
		return stats, err
	}

	rows, err := s.Pool.Query(ctx,
		`SELECT id, email, display_name, role, created_at
		 FROM members ORDER BY created_at DESC, id DESC LIMIT 5`)
	if err != nil {
		return stats, err
	}
	stats.NewestMembers = make([]Member, 0)
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.ID, &member.Email, &member.DisplayName, &member.Role, &member.CreatedAt); err != nil {
			rows.Close()
			return stats, err
		}
		stats.NewestMembers = append(stats.NewestMembers, member)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return stats, err
	}
	rows.Close()

	rows, err = s.Pool.Query(ctx,
		`SELECT m.id, m.title, m.title_th, `+posterSQL+`,
		        vs.vote_count, vs.vote_average::double precision
		 FROM movie_vote_stats vs
		 JOIN movies m ON m.id = vs.movie_id
		 ORDER BY vs.vote_count DESC, m.id
		 LIMIT 5`)
	if err != nil {
		return stats, err
	}
	stats.TopVoted = make([]TopMovie, 0)
	defer rows.Close()
	for rows.Next() {
		var movie TopMovie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.TitleTh, &movie.Poster, &movie.VoteCount, &movie.VoteAverage); err != nil {
			return stats, err
		}
		stats.TopVoted = append(stats.TopVoted, movie)
	}
	if err := rows.Err(); err != nil {
		return stats, err
	}
	return stats, nil
}

func (s *Store) ListGenres(ctx context.Context) ([]Genre, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, name_en, name_th FROM genres ORDER BY name_en`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	genres := make([]Genre, 0)
	for rows.Next() {
		var genre Genre
		if err := rows.Scan(&genre.ID, &genre.Name, &genre.NameTH); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return genres, nil
}

type MovieInput struct {
	Title            string  `json:"title"`
	TitleTh          *string `json:"titleTh"`
	OriginalTitle    string  `json:"originalTitle"`
	OriginalLanguage string  `json:"originalLanguage"`
	Overview         *string `json:"overview"`
	OverviewTh       *string `json:"overviewTh"`
	PosterPath       *string `json:"posterPath"`
	ReleaseDate      *string `json:"releaseDate"`
	GenreID          *int    `json:"genreId"`
}

type Impact struct {
	Reviews  int64 `json:"reviews"`
	Votes    int64 `json:"votes"`
	Wishlist int64 `json:"wishlist"`
}

type AdminMovie struct {
	ID int64 `json:"id"`
	MovieInput
	Poster      *string   `json:"poster"`
	VoteAverage float64   `json:"voteAverage"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Impact      Impact    `json:"impact"`
}

func (s *Store) AdminGetMovie(ctx context.Context, id int64) (AdminMovie, error) {
	var movie AdminMovie
	var releaseDate *time.Time
	err := s.Pool.QueryRow(ctx,
		`SELECT m.id, m.title, m.title_th, m.original_title, m.original_language,
		        m.overview, m.overview_th, m.poster_path, m.release_date,
		        (SELECT genre_id FROM movie_genres WHERE movie_id = m.id AND position = 0),
		        `+posterSQL+`, round(m.vote_average, 1)::double precision, m.updated_at,
		        (SELECT count(*) FROM reviews WHERE movie_id = m.id),
		        (SELECT count(*) FROM votes WHERE movie_id = m.id),
		        (SELECT count(*) FROM wishlist_items WHERE movie_id = m.id)
		 FROM movies m WHERE m.id = $1`, id).
		Scan(&movie.ID, &movie.Title, &movie.TitleTh, &movie.OriginalTitle, &movie.OriginalLanguage,
			&movie.Overview, &movie.OverviewTh, &movie.PosterPath, &releaseDate, &movie.GenreID,
			&movie.Poster, &movie.VoteAverage, &movie.UpdatedAt,
			&movie.Impact.Reviews, &movie.Impact.Votes, &movie.Impact.Wishlist)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminMovie{}, ErrNotFound
	}
	if releaseDate != nil {
		date := releaseDate.Format("2006-01-02")
		movie.ReleaseDate = &date
	}
	return movie, err
}

func setMainGenre(ctx context.Context, tx pgx.Tx, movieID int64, genreID *int) error {
	if genreID == nil {
		_, err := tx.Exec(ctx, `DELETE FROM movie_genres WHERE movie_id = $1 AND position = 0`, movieID)
		return err
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM movie_genres WHERE movie_id = $1 AND (position = 0 OR genre_id = $2)`,
		movieID, *genreID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO movie_genres (movie_id, genre_id, position) VALUES ($1, $2, 0)`,
		movieID, *genreID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrBadGenre
	}
	return err
}

func (s *Store) CreateMovie(ctx context.Context, id int64, input MovieInput) (AdminMovie, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return AdminMovie{}, err
	}
	defer tx.Rollback(ctx)

	if id == 0 {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(913775)`); err != nil {
			return AdminMovie{}, err
		}
		if err := tx.QueryRow(ctx,
			`SELECT greatest(coalesce(max(id), 0) + 1, 9000000) FROM movies`).Scan(&id); err != nil {
			return AdminMovie{}, err
		}
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO movies (id, title, title_th, original_title, original_language,
		                     overview, overview_th, poster_path, release_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date)`,
		id, input.Title, input.TitleTh, input.OriginalTitle, input.OriginalLanguage,
		input.Overview, input.OverviewTh, input.PosterPath, input.ReleaseDate)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return AdminMovie{}, ErrDuplicateID
	}
	if err != nil {
		return AdminMovie{}, err
	}
	if err := setMainGenre(ctx, tx, id, input.GenreID); err != nil {
		return AdminMovie{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminMovie{}, err
	}
	return s.AdminGetMovie(ctx, id)
}

func (s *Store) UpdateMovie(ctx context.Context, id int64, input MovieInput) (AdminMovie, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return AdminMovie{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE movies
		 SET title = $2, title_th = $3, original_title = $4, original_language = $5,
		     overview = $6, overview_th = $7, poster_path = $8, release_date = $9::date,
		     updated_at = now()
		 WHERE id = $1`,
		id, input.Title, input.TitleTh, input.OriginalTitle, input.OriginalLanguage,
		input.Overview, input.OverviewTh, input.PosterPath, input.ReleaseDate)
	if err != nil {
		return AdminMovie{}, err
	}
	if tag.RowsAffected() == 0 {
		return AdminMovie{}, ErrNotFound
	}
	if err := setMainGenre(ctx, tx, id, input.GenreID); err != nil {
		return AdminMovie{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminMovie{}, err
	}
	return s.AdminGetMovie(ctx, id)
}

func (s *Store) DeleteMovie(ctx context.Context, id int64) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM movies WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
