package mastodon

import (
	"fmt"
	"net/http"

	"github.com/davecheney/pub/internal/algorithms"
	"github.com/davecheney/pub/internal/httpx"
	"github.com/davecheney/pub/internal/models"
	"github.com/davecheney/pub/internal/to"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func BookmarksIndex(env *Env, w http.ResponseWriter, r *http.Request) error {
	user, err := env.authenticate(r)
	if err != nil {
		return err
	}

	var bookmarked []*models.Status
	query := env.DB.Joins("JOIN reactions ON reactions.status_id = statuses.id and reactions.actor_id = ? and reactions.bookmarked = ?", user.Actor.ID, true)
	query = query.Preload("Actor")
	query = query.Scopes(models.PreloadStatus, models.PreloadReaction(user.Actor))
	if err := query.Find(&bookmarked).Error; err != nil {
		return err
	}

	w.Header().Add("Link", fmt.Sprintf(`<https://%s/api/v1/bookmarks?min_id=%d>; rel="prev"`, r.Host, bookmarked[len(bookmarked)-1].ID))
	return to.JSON(w, algorithms.Map(bookmarked, serialiseStatus))
}

func BookmarksCreate(env *Env, w http.ResponseWriter, r *http.Request) error {
	user, err := env.authenticate(r)
	if err != nil {
		return err
	}
	var status models.Status
	query := env.DB.Joins("Actor").Scopes(models.PreloadStatus, models.PreloadReaction(user.Actor))
	if err := query.Take(&status, chi.URLParam(r, "id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return httpx.Error(http.StatusNotFound, err)
		}
		return err
	}
	reaction, err := models.NewReactions(env.DB).Bookmark(&status, user.Actor)
	if err != nil {
		return err
	}
	return to.JSON(w, serialiseStatus(reaction.Status))
}

func BookmarksDestroy(env *Env, w http.ResponseWriter, r *http.Request) error {
	user, err := env.authenticate(r)
	if err != nil {
		return err
	}
	var status models.Status
	query := env.DB.Joins("Actor").Scopes(models.PreloadStatus, models.PreloadReaction(user.Actor))
	if err := query.Take(&status, chi.URLParam(r, "id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return httpx.Error(http.StatusNotFound, err)
		}
		return err
	}
	reaction, err := models.NewReactions(env.DB).Unbookmark(&status, user.Actor)
	if err != nil {
		return err
	}
	return to.JSON(w, serialiseStatus(reaction.Status))
}
