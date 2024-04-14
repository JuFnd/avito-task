package delivery

import (
	"avito-track/pkg/middleware"
	"avito-track/pkg/models"
	"avito-track/pkg/util"
	"avito-track/pkg/variables"
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type ICore interface {
	UserBanner(tagID int64, featureID int64, useLastRevision bool) (*models.Banner, error)
	GetBanners(featureID int64, tagIDs []int64, limit, offset int64) ([]models.Banner, error)
	GetUserId(ctx context.Context, sid string) (int64, error)
	GetUserRole(ctx context.Context, id int64) (string, error)
}

type API struct {
	core   ICore
	logger *slog.Logger
	mux    *http.ServeMux
}

func (api *API) ListenAndServe(appConfig *variables.AppConfig) error {
	err := http.ListenAndServe(appConfig.Address, api.mux)
	if err != nil {
		api.logger.Error(variables.ListenAndServeError, err.Error())
		return err
	}
	return nil
}

func GetApi(bannerCore ICore, bannerLogger *slog.Logger) *API {
	api := &API{
		core:   bannerCore,
		logger: bannerLogger,
		mux:    http.NewServeMux(),
	}

	api.mux.Handle("/api/v1/user_banner", middleware.MethodMiddleware(
		middleware.AuthorizationMiddleware(
			http.HandlerFunc(api.UserBanner),
			api.core,
			api.logger),
		http.MethodGet, api.logger))

	api.mux.Handle("/api/v1/banner", middleware.MethodMiddleware(
		middleware.AuthorizationMiddleware(
			middleware.PermissionsMiddleware(
				http.HandlerFunc(api.BannersList),
				api.core,
				variables.AdminRole,
				api.logger),
			api.core,
			api.logger),
		http.MethodGet, api.logger))

	api.mux.Handle("/api/v1/banner/", middleware.MethodMiddleware(
		middleware.AuthorizationMiddleware(
			middleware.PermissionsMiddleware(
				http.HandlerFunc(api.BannersSettings),
				api.core,
				variables.AdminRole,
				api.logger),
			api.core,
			api.logger),
		http.MethodGet, api.logger))
	return api
}

func (api *API) UserBanner(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.URL.Query().Get("tag_id")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil || tagID <= 0 {
		util.SendResponse(w, r, http.StatusBadRequest, nil, variables.TagIdError, err, api.logger)
		return
	}

	featureIDStr := r.URL.Query().Get("feature_id")
	featureID, err := strconv.ParseInt(featureIDStr, 10, 64)
	if err != nil || featureID <= 0 {
		util.SendResponse(w, r, http.StatusBadRequest, nil, variables.FeatureIdError, err, api.logger)
		return
	}

	useLastRevisionStr := r.URL.Query().Get("use_last_revision")
	var useLastRevision bool
	if useLastRevisionStr == "" {
		useLastRevision = false
	} else {
		useLastRevision, err = strconv.ParseBool(useLastRevisionStr)
		if err != nil {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.LastRevisionError, err, api.logger)
			return
		}
	}

	banner, err := api.core.UserBanner(tagID, featureID, useLastRevision)
	if err != nil {
		util.SendResponse(w, r, http.StatusNotFound, nil, variables.BannerNotFoundError, err, api.logger)
		return
	}

	util.SendResponse(w, r, http.StatusOK, banner, variables.StatusOkMessage, nil, api.logger)
}

func (api *API) BannersList(w http.ResponseWriter, r *http.Request) {
	featureIDStr := r.URL.Query().Get("feature_id")
	var featureID int64
	if featureIDStr != "" {
		var err error
		featureID, err = strconv.ParseInt(featureIDStr, 10, 64)
		if err != nil || featureID < 1 {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.FeatureIdError, err, api.logger)
			return
		}
	}

	tagIDStr := r.URL.Query().Get("tag_id")
	var tagIDs []int64
	if tagIDStr != "" {
		tags := strings.Split(tagIDStr, ",")
		for _, rawTagID := range tags {
			id, err := strconv.ParseInt(rawTagID, 10, 64)
			if err != nil || id < 1 {
				util.SendResponse(w, r, http.StatusBadRequest, nil, variables.TagIdError, err, api.logger)
				return
			}
			tagIDs = append(tagIDs, id)
		}
	}

	limitStr := r.URL.Query().Get("limit")
	var limit int64
	if limitStr != "" {
		lim, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || lim < 1 {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.InvalidLimit, err, api.logger)
			return
		}
		limit = lim
	} else {
		limit = 10
	}

	offsetStr := r.URL.Query().Get("offset")
	var offset int64
	if offsetStr != "" {
		off, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil || off < 0 {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.InvalidOffset, err, api.logger)
			return
		}
		offset = off
	} else {
		offset = 0
	}

	banners, err := api.core.GetBanners(featureID, tagIDs, limit, offset)
	if err != nil {
		util.SendResponse(w, r, http.StatusInternalServerError, nil, variables.StatusInternalServerError, err, api.logger)
		return
	}

	util.SendResponse(w, r, http.StatusOK, banners, variables.StatusOkMessage, nil, api.logger)
}

func (api *API) BannersSettings(w http.ResponseWriter, r *http.Request) {
	//implement
}
