package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

func (h *Handler) AdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Store.Stats(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) AdminMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.Store.ListMembers(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": members})
}

type roleReq struct {
	Role string `json:"role"`
}

func (h *Handler) AdminSetRole(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสมาชิกต้องเป็นตัวเลข")
		return
	}
	var req roleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if req.Role != "member" && req.Role != "admin" {
		writeError(w, http.StatusBadRequest, "BAD_ROLE", "role ต้องเป็น member หรือ admin")
		return
	}
	if memberID, _ := mw.MemberID(r.Context()); memberID == id {
		writeError(w, http.StatusBadRequest, "SELF_ROLE", "เปลี่ยนบทบาทของตัวเองไม่ได้ ให้ admin คนอื่นทำแทน")
		return
	}
	member, err := h.Store.UpdateMemberRole(r.Context(), id, req.Role)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "ไม่พบสมาชิกคนนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

func (h *Handler) Genres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.Store.ListGenres(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": genres})
}

type movieReq struct {
	ID int64 `json:"id"`
	store.MovieInput
}

func trimPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validateMovie(movie *store.MovieInput) (string, string) {
	movie.Title = strings.TrimSpace(movie.Title)
	movie.OriginalTitle = strings.TrimSpace(movie.OriginalTitle)
	movie.OriginalLanguage = strings.ToLower(strings.TrimSpace(movie.OriginalLanguage))
	movie.TitleTh = trimPtr(movie.TitleTh)
	movie.Overview = trimPtr(movie.Overview)
	movie.OverviewTh = trimPtr(movie.OverviewTh)
	movie.PosterPath = trimPtr(movie.PosterPath)
	movie.ReleaseDate = trimPtr(movie.ReleaseDate)

	switch {
	case movie.Title == "" || utf8.RuneCountInString(movie.Title) > 255:
		return "BAD_TITLE", "ต้องมีชื่อเรื่อง และยาวไม่เกิน 255 ตัวอักษร"
	case movie.TitleTh != nil && utf8.RuneCountInString(*movie.TitleTh) > 255:
		return "BAD_TITLE_TH", "ชื่อภาษาไทยยาวไม่เกิน 255 ตัวอักษร"
	case movie.OriginalTitle == "":
		movie.OriginalTitle = movie.Title
	case utf8.RuneCountInString(movie.OriginalTitle) > 255:
		return "BAD_ORIGINAL_TITLE", "ชื่อเรื่องต้นฉบับยาวไม่เกิน 255 ตัวอักษร"
	}
	if movie.OriginalLanguage == "" {
		movie.OriginalLanguage = "en"
	}
	if len(movie.OriginalLanguage) != 2 {
		return "BAD_LANGUAGE", "รหัสภาษาต้องเป็น 2 ตัวอักษร เช่น en, th, ja"
	}
	if movie.PosterPath != nil {
		poster := *movie.PosterPath
		if len(poster) > 500 {
			return "BAD_POSTER", "ลิงก์โปสเตอร์ยาวไม่เกิน 500 ตัวอักษร"
		}
		if !strings.HasPrefix(poster, "/") {
			parsed, err := url.ParseRequestURI(poster)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return "BAD_POSTER", "โปสเตอร์ต้องเป็น path ของ TMDB (/xxx.jpg) หรือ URL เต็ม (https://...)"
			}
		}
	}
	if movie.ReleaseDate != nil {
		if _, err := time.Parse("2006-01-02", *movie.ReleaseDate); err != nil {
			return "BAD_DATE", "วันฉายต้องอยู่ในรูปแบบ YYYY-MM-DD"
		}
	}
	return "", ""
}

func (h *Handler) AdminGetMovie(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	movie, err := h.Store.AdminGetMovie(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

func (h *Handler) AdminCreateMovie(w http.ResponseWriter, r *http.Request) {
	var req movieReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if req.ID < 0 {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็น 0 (ให้ระบบออก) หรือเลขบวก")
		return
	}
	if code, message := validateMovie(&req.MovieInput); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}
	movie, err := h.Store.CreateMovie(r.Context(), req.ID, req.MovieInput)
	switch {
	case errors.Is(err, store.ErrDuplicateID):
		writeError(w, http.StatusConflict, "MOVIE_EXISTS", "มีหนังเลขนี้อยู่แล้ว")
		return
	case errors.Is(err, store.ErrBadGenre):
		writeError(w, http.StatusBadRequest, "BAD_GENRE", "ไม่มีแนวหนังรหัสนี้")
		return
	case err != nil:
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, movie)
}

func (h *Handler) AdminUpdateMovie(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req movieReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if code, message := validateMovie(&req.MovieInput); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}
	movie, err := h.Store.UpdateMovie(r.Context(), id, req.MovieInput)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	case errors.Is(err, store.ErrBadGenre):
		writeError(w, http.StatusBadRequest, "BAD_GENRE", "ไม่มีแนวหนังรหัสนี้")
		return
	case err != nil:
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

func (h *Handler) AdminDeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	if err := h.Store.DeleteMovie(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	} else if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
