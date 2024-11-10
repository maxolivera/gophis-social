package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/models"
)

func (app *Application) handlerSearch(w http.ResponseWriter, r *http.Request) {
	/*
		Filter parameters:
		* Tags
		* Search (fuzzy search)
		* Since
	*/
	params := database.SearchPostsParams{
		Search: "",
		Tags:   nil,
		Limit:  10,    // Default limit
		Offset: 0,     // Default offset
		Sort:   false, // Default sort order
	}
	ctx := r.Context()
	url := r.URL.Query()

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	// tags
	tagsStr := url.Get("tags")
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		if tags != nil {
			params.Tags = tags
		}
	}

	// search
	search := url.Get("search")
	if search != "" {
		params.Search = search
	}

	// dates
	sinceStr := url.Get("since")
	var since time.Time
	if sinceStr != "" {
		var err error
		since, err = time.Parse("2006-01-02", sinceStr)
		if err != nil {
			log.Printf("could not parse date: %v\n", err)
		}
		params.Since = pgtype.Timestamp{Time: since}
	}
	untilStr := url.Get("until")
	var until time.Time
	if untilStr != "" {
		var err error
		until, err = time.Parse("2006-01-02", untilStr)
		if err != nil {
			log.Printf("could not parse date: %v\n", err)
		}
		params.Until = pgtype.Timestamp{Time: until}
	}

	// limit and offset
	limitStr := url.Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			log.Printf("could not parse limit into int: %v", err)
		} else {
			if limit >= 20 {
				limit = 20
			}
			params.Limit = int32(limit)
		}
	} else {
		params.Limit = 20
	}
	offsetStr := url.Get("offset")
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Printf("could not parse limit into int: %v", err)
		} else {
			params.Offset = int32(offset)
		}
	} else {
		params.Offset = 0
	}

	// sort
	sort := url.Get("sort")
	if sort != "desc" && sort != "asc" {
		// asume desc
		params.Sort = true
	} else {
		if sort == "desc" {
			params.Sort = true
		} else {
			params.Sort = false
		}
	}

	dbFeed, err := app.Database.SearchPosts(ctx, params)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			err = fmt.Errorf("no rows: %v", err)
			respondWithError(w, r, http.StatusNotFound, err, "no posts with these parameters")
		default:
			err = fmt.Errorf("error when retrieving posts with filter parameters %v, err: %v", params, err)
			respondWithError(w, r, http.StatusInternalServerError, err, "")
		}
		return
	}

	feed, err := models.DBFeedsToFeeds(dbFeed)
	if err != nil {
		err = fmt.Errorf("error during parsing: %v", err)
		respondWithError(w, r, http.StatusInternalServerError, err, "")
		return
	}

	respondWithJSON(w, http.StatusOK, feed)
}
