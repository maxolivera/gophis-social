package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/maxolivera/gophis-social-network/internal/storage"
)

// Search godoc
//
//	@Summary		Search posts
//	@Description	Search posts according to parameters.
//	@tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			search	path		string	false	"Search both in Post's Title and Content"
//	@Param			tags	path		string	false	"Tags"
//	@Param			limit	path		int32	false	"Number of posts. Default 10; Maximum 20"
//	@Param			offset	path		int32	false	"Offset. Default at 0"
//	@Param			sort	path		bool	false	"Sort, true if descending order"
//	@Success		200		{object}	[]models.Feed
//	@Failure		401		{object}	error "Unauthorized"
//	@Failure		404		{object}	error "No posts with selected parameters"
//	@Failure		500		{object}	error
//	@Router			/v1/search [get]
//	@Security		ApiKeyAuth
func (app *Application) handlerSearch(w http.ResponseWriter, r *http.Request) {
	/*
		Filter parameters:
		* Tags
		* Search (fuzzy search)
		* Since
	*/
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	url := r.URL.Query()

	var word string
	var tags []string
	var limit int32
	var offset int32
	var sort bool
	var since *time.Time
	var until *time.Time

	// tags
	tagsStr := url.Get("tags")
	if tagsStr != "" {
		splittedTags := strings.Split(tagsStr, ",")
		if tags != nil {
			tags = splittedTags
		}
	}

	// search
	search := url.Get("search")
	if search != "" {
		word = search
	}

	// dates
	sinceStr := url.Get("since")
	if sinceStr != "" {
		sinceDate, err := time.Parse("2006-01-02", sinceStr)
		if err != nil {
			err = fmt.Errorf("could not parse date: %v\n", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			return
		}
		since = &sinceDate
	}

	untilStr := url.Get("until")
	if untilStr != "" {
		untilDate, err := time.Parse("2006-01-02", untilStr)
		if err != nil {
			err = fmt.Errorf("could not parse date: %v\n", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			return
		}
		until = &untilDate
	}

	// limit and offset
	limitStr := url.Get("limit")
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil {
			err = fmt.Errorf("could not parse limit into int: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			return
		} else {
			if limit >= 20 {
				limit = 20
			}
			limit = int32(limitInt)
		}
	} else {
		limit = -1
	}

	offsetStr := url.Get("offset")
	if offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil {
			err = fmt.Errorf("could not parse limit into int: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			return
		} else {
			offset = int32(offsetInt)
		}
	} else {
		offset = -1
	}

	// sort
	sortStr := url.Get("sort")
	if sortStr != "desc" && sortStr != "asc" {
		// asume desc
		sort = true
	} else {
		if sortStr == "desc" {
			sort = true
		} else {
			sort = false
		}
	}

	feed, err := app.Storage.Posts.Search(
		ctx, word, tags, limit, offset, sort, since, until,
	)
	if err != nil {
		switch err {
		case storage.ErrNoRows:
			err = fmt.Errorf("no rows: %v", err)
			app.respondWithError(w, r, http.StatusNotFound, err, "no posts with these parameters")
		default:
			err = fmt.Errorf("error when retrieving posts with filter parameters, err: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	app.respondWithJSON(w, r, http.StatusOK, feed)
}
