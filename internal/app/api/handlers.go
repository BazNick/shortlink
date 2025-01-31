package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
)

func AddLink(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, apperr.ErrOnlyPOST, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, apperr.ErrBodyRead, http.StatusBadRequest)
		return
	}

	alreadyExst := entities.CheckValExists(entities.Hash, string(body))
	if alreadyExst {
		http.Error(res, apperr.ErrLinkExists, http.StatusBadRequest)
		return
	}

	var scheme string
    if req.TLS != nil {
        scheme = "https://"
    } else {
        scheme = "http://"
    }

	var (
		randStr  = functions.RandSeq(8)
		hashLink = scheme + req.Host + "/" + randStr
	)

	entities.Hash.AddHash(randStr, string(body))

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(hashLink))
}

func GetLink(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, apperr.ErrOnlyGET, http.StatusMethodNotAllowed)
		return
	}

	var (
		id     = strings.TrimPrefix(req.URL.Path, "/")
		exists = entities.Hash.GetHash(id)
	)

	if exists == "" {
		http.Error(res, apperr.ErrLinkNotFound, http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", exists)
	res.WriteHeader(http.StatusTemporaryRedirect)
}