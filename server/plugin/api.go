package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	toChannelKey      = "to_channel"
	additionalTextKey = "additional_text"
)

var messageGenericError = toPtr("Something went wrong. Please try again later.")

type submitDialogHandler func(map[string]string, *model.SubmitDialogRequest) (*string, *model.SubmitDialogResponse, error)

func (p *SharePostPlugin) InitAPI() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", p.handleInfo).Methods(http.MethodGet)

	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.Use(checkAuthenticity)
	apiV1.HandleFunc("/share", p.handleSubmitDialogRequest(p.handleSharePost)).Methods(http.MethodPost)
	// apiV1.HandleFunc("/move", p.handleSubmitDialogRequest(p.handleMovePost).Methods(http.MethodPost)
	return r
}

func (p *SharePostPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.router.ServeHTTP(w, r)
}

func (p *SharePostPlugin) handleInfo(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, fmt.Sprintf("Installed SharePostPlugin v%s", manifest.Version))
}

func checkAuthenticity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Mattermost-User-ID") == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (p *SharePostPlugin) handleSubmitDialogRequest(handler submitDialogHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := model.SubmitDialogRequestFromJson(r.Body)
		if request == nil {
			p.API.LogWarn("Failed to decode SubmitDialogRequest")
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if request.UserId != r.Header.Get("Mattermost-User-Id") {
			p.API.LogWarn("invalid user")
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		msg, response, err := handler(mux.Vars(r), request)
		if err != nil {
			p.API.LogWarn("Failed to handle SubmitDialogRequest", "error", err.Error())
		}

		if msg != nil {
			p.SendEphemeralPost(request.ChannelId, request.UserId, *msg)
		}

		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				p.API.LogWarn("Failed to write SubmitDialogRequest", "error", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}

func (p *SharePostPlugin) handleSharePost(vars map[string]string, request *model.SubmitDialogRequest) (*string, *model.SubmitDialogResponse, error) {
	postId := request.CallbackId
	teamId := request.TeamId

	team, appErr := p.API.GetTeam(teamId)
	if appErr != nil {
		p.API.LogError("Failed to get team", "team_id", teamId, "error", appErr.Error())
		return messageGenericError, nil, fmt.Errorf("Failed to get team %w", appErr)
	}

	toChannel, ok := request.Submission[toChannelKey].(string)
	if !ok {
		return messageGenericError, nil, errors.Errorf("failed to get toChannel key. Value is: %v", request.Submission[toChannelKey])
	}
	additionalText, ok := request.Submission[additionalTextKey].(string)
	if ok {
		additionalText = fmt.Sprintf("%s\n\n", additionalText)
	}

	postLink := fmt.Sprintf("%s/%s/pl/%s", *p.ServerConfig.ServiceSettings.SiteURL, team.Name, postId)
	if _, err := p.API.CreatePost(&model.Post{
		Type:      model.POST_DEFAULT,
		UserId:    request.UserId,
		ChannelId: toChannel,
		Message:   fmt.Sprintf("%s> Shared from %s", additionalText, postLink),
	}); err != nil {
		p.API.LogWarn("Failed to create post", "error", err.Error())
		return messageGenericError, nil, fmt.Errorf("Failed to create post %w", err)
	}
	return nil, nil, nil
}

func toPtr(s string) *string {
	return &s
}
