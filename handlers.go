package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/koheiyamayama/google-cloud-go/models"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type (
	Handlers struct {
		mysqlClient *models.MySQLClient
	}

	ErrorResponse struct {
		Msg string `json:"msg"`
	}
)

func NewHandlers(mysqlClient *models.MySQLClient) *Handlers {
	return &Handlers{
		mysqlClient: mysqlClient,
	}
}

func (h *Handlers) ListPosts(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("list posts")
	w.Header().Add("Content-Type", "application/json")
	ctx := r.Context()
	posts, err := h.mysqlClient.ListPosts(ctx, models.ToPtr(10))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errRes := &ErrorResponse{
			Msg: err.Error(),
		}
		b, _ := json.Marshal(errRes)
		w.Write(b)
		return
	}
	log.Debug().Msgf("found posts")
	type postsResponse struct {
		Posts []*models.PostWithUser `json:"posts"`
	}

	for _, post := range posts {
		post.User.ProfileURL = post.User.ProfileImageURL()
	}

	postsRes := &postsResponse{
		Posts: posts,
	}
	b, err := json.Marshal(postsRes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errRes := &ErrorResponse{
			Msg: err.Error(),
		}
		b, _ := json.Marshal(errRes)
		w.Write(b)
		return
	}
	log.Debug().Msgf("response posts")
	w.Write(b)
}

func (h *Handlers) CreatePosts(w http.ResponseWriter, r *http.Request) {
	b := r.Body
	defer r.Body.Close()
	v := &models.Post{}
	if err := json.NewDecoder(b).Decode(v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	retVal, err := h.mysqlClient.InsertPost(ctx, v.Title, v.Body, v.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errRes := &ErrorResponse{
			Msg: err.Error(),
		}
		b, _ := json.Marshal(errRes)
		w.Write(b)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(retVal.String()))
}

func (h *Handlers) ListPostsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParamFromCtx(ctx, "userID")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		errRes := &ErrorResponse{
			Msg: "must specified userID",
		}
		b, _ := json.Marshal(errRes)
		w.Write(b)
		return
	}

	posts, err := h.mysqlClient.SelectPostsByUserID(ctx, ulid.MustParse(userID), models.ToPtr(20))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errRes := &ErrorResponse{
			Msg: err.Error(),
		}
		b, _ := json.Marshal(errRes)
		w.Write(b)
		return
	}

	type postsResponse struct {
		Posts []*models.Post `json:"posts"`
	}

	postsRes := postsResponse{
		Posts: posts,
	}

	b, err := json.Marshal(postsRes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errRes := &ErrorResponse{
			Msg: err.Error(),
		}
		b, _ := json.Marshal(errRes)
		w.Write(b)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
