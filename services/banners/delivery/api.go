package delivery

import (
	"avito-track/pkg/middleware"
	"avito-track/pkg/models"
	communication "avito-track/pkg/requests"
	"avito-track/pkg/util"
	"avito-track/pkg/variables"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type ICore interface {
	UserBanner(tagID int64, featureID int64, useLastRevision bool) (*models.Banner, error)
	GetBanners(userRole string, featureID int64, tagIDs []int64, limit, offset int64) ([]models.Banner, error)
	AddBanner(tagIDs []int64, featureID int64, content string) error
	UpdateBanner(id int64, tagIds []int64, featureID int64, content string) error
	DeleteBanner(id int64) error
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
		variables.MethodGet, api.logger))

	api.mux.Handle("/api/v1/banner", middleware.MethodMiddleware(
		middleware.AuthorizationMiddleware(
			middleware.PermissionsMiddleware(
				http.HandlerFunc(api.BannersList),
				api.core,
				variables.AdminAndUser,
				api.logger),
			api.core,
			api.logger),
		variables.MethodGetAndPost, api.logger))

	api.mux.Handle("/api/v1/banner/", middleware.MethodMiddleware(
		middleware.AuthorizationMiddleware(
			middleware.PermissionsMiddleware(
				http.HandlerFunc(api.BannersSettings),
				api.core,
				variables.AdminRole,
				api.logger),
			api.core,
			api.logger),
		variables.MethodsDeletePatch, api.logger))
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
	userRole, isRole := r.Context().Value(variables.RoleKey).(string)
	if !isRole {
		util.SendResponse(w, r, http.StatusUnauthorized, nil, variables.StatusUnauthorizedError, nil, api.logger)
		return
	}

	switch r.Method {
	case http.MethodGet:
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

		banners, err := api.core.GetBanners(userRole, featureID, tagIDs, limit, offset)
		if err != nil {
			util.SendResponse(w, r, http.StatusInternalServerError, nil, variables.StatusInternalServerError, err, api.logger)
			return
		}

		util.SendResponse(w, r, http.StatusOK, banners, variables.StatusOkMessage, nil, api.logger)
	case http.MethodPost:
		if userRole != variables.AdminRole[0] {
			util.SendResponse(w, r, http.StatusForbidden, nil, variables.StatusForbiddenError, nil, api.logger)
			return
		}

		var banner communication.BannerRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.StatusBadRequestError, nil, api.logger)
			return
		}

		err = json.Unmarshal(body, &banner)
		if err != nil {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.StatusBadRequestError, nil, api.logger)
			return
		}

		err = api.core.AddBanner(banner.TagIds, banner.FeatureId, banner.Content)
		if err != nil {
			util.SendResponse(w, r, http.StatusInternalServerError, nil, variables.StatusInternalServerError, nil, api.logger)
			return
		}

		util.SendResponse(w, r, http.StatusCreated, nil, variables.StatusOkMessage, nil, api.logger)
	}
}

func (api *API) BannersSettings(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/banner/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		util.SendResponse(w, r, http.StatusBadRequest, nil, variables.StatusBadRequestError, err, api.logger)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var banner communication.BannerRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			util.SendResponse(w, r, http.StatusBadRequest, nil, variables.StatusBadRequestError, err, api.logger)
			return
		}

		err = json.Unmarshal(body, &banner)
		if err != nil {
			util.SendResponse(w, r, http.StatusInternalServerError, nil, variables.StatusInternalServerError, err, api.logger)
			return
		}

		err = api.core.UpdateBanner(id, banner.TagIds, banner.FeatureId, banner.Content)
		if err != nil {
			util.SendResponse(w, r, http.StatusNotFound, nil, variables.BannerNotFoundError, err, api.logger)
			return
		}

		util.SendResponse(w, r, http.StatusOK, nil, variables.StatusOkMessage, nil, api.logger)
	case http.MethodDelete:
		err := api.core.DeleteBanner(id)
		if err != nil {
			util.SendResponse(w, r, http.StatusInternalServerError, nil, variables.BannerNotFoundError, err, api.logger)
			return
		}
		util.SendResponse(w, r, http.StatusOK, nil, variables.StatusOkMessage, nil, api.logger)
	}
}
