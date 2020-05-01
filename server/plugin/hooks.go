package plugin

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func (p *SharePostPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
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
		oldPostCreateAt := time.Unix(oldPost.CreateAt/1000, 0)
		attachment := []*model.SlackAttachment{
			{
				Timestamp:  oldPost.CreateAt,
				AuthorName: postUser.GetDisplayNameWithPrefix(model.SHOW_NICKNAME_FULLNAME, "@"),
				Text:       oldPost.Message,
				Footer: fmt.Sprintf("Posted in ~%s %s",
					oldchannel.Name,
					oldPostCreateAt.Format("on Mon 2 Jan 2006 at 15:04:05 MST"),
				),
			},
			nil,
		}
		model.ParseSlackAttachment(post, attachment)
	}

	return post, ""

}
