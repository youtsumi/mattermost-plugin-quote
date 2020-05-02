package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	// https://pkg.go.dev/github.com/mattermost/mattermost-server/v5/model
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	channel, err := p.API.GetChannel(post.ChannelId)
	if err != nil {
		return post, err.Message
	}

	team, err := p.API.GetTeam(channel.TeamId)
	if err != nil {
		return post, err.Message
	}

	selfLink := fmt.Sprintf("%s/%s", *siteURL, team.Name)
	selfLinkPattern, er := regexp.Compile(fmt.Sprintf("%s%s", selfLink, `/[\w/]+`))
	if er != nil {
		return post, er.Error()
	}

	for _, match := range selfLinkPattern.FindAllString(post.Message, -1) {
		separated := strings.Split(match, "/")
		postId := separated[len(separated)-1]
		oldPost, err := p.API.GetPost(postId)
		if err != nil {
			return post, err.Message
		}

		oldchannel, err := p.API.GetChannel(oldPost.ChannelId)
		if err != nil {
			return post, err.Message
		}

		postUser, err := p.API.GetUser(oldPost.UserId)
		if err != nil {
			return post, err.Message
		}

		oldPostCreateAt := time.Unix(oldPost.CreateAt/1000,0)
		attachment := []*model.SlackAttachment{
			{
				Timestamp:  oldPost.CreateAt,
				AuthorName: postUser.GetDisplayNameWithPrefix(model.SHOW_NICKNAME_FULLNAME,"@"),
				AuthorIcon: fmt.Sprintf("%s/api/v4/users/%s/image", *siteURL, oldPost.UserId),
				Text:       oldPost.Message,
				Footer:     fmt.Sprintf("Posted in ~%s %s",
						oldchannel.Name,
						oldPostCreateAt.Format("on Mon 2 Jan 2006 at 15:04:05 MST"),
					),
			},
			nil,
		}
		model.ParseSlackAttachment(post,attachment)
	}

	return post, ""

}

// See https://developers.mattermost.com/extend/plugins/server/reference/
