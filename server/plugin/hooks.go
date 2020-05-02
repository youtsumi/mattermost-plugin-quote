package plugin

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// MessageWillBePosted expand contents of permalink of local post
func (p *SharePostPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	channel, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil {
		return post, appErr.Error()
	}

	team, appErr := p.API.GetTeam(channel.TeamId)
	if appErr != nil {
		return post, appErr.Error()
	}

	selfLink := fmt.Sprintf("%s/%s", *siteURL, team.Name)
	selfLinkPattern, err := regexp.Compile(fmt.Sprintf("%s%s", selfLink, `/[\w/]+`))
	if err != nil {
		return post, err.Error()
	}

	// Only first post matched the pattern is expanded
	for _, match := range selfLinkPattern.FindAllString(post.Message, 1) {
		separated := strings.Split(match, "/")
		postID := separated[len(separated)-1]
		oldPost, appErr := p.API.GetPost(postID)
		if appErr != nil {
			return post, appErr.Error()
		}

		newFileIds, appErr := p.API.CopyFileInfos(post.UserId, oldPost.FileIds)
		if appErr != nil {
			p.API.LogWarn("Failed to copy file ids", "error", appErr.Error())
			return post, appErr.Error()
		}
		// CAUTION: if attaching over 5 files, error will occur
		post.FileIds = append(post.FileIds, newFileIds...)

		oldchannel, appErr := p.API.GetChannel(oldPost.ChannelId)
		if appErr != nil {
			return post, appErr.Error()
		}

		postUser, appErr := p.API.GetUser(oldPost.UserId)
		if appErr != nil {
			return post, appErr.Error()
		}
		oldPostCreateAt := time.Unix(oldPost.CreateAt/1000, 0)
		ateAt := time.Unix(oldPost.CreateAt/1000,0)

		AuthorName :=postUser.GetDisplayNameWithPrefix(model.SHOW_NICKNAME_FULLNAME,"@")
		fmtstmnt := "%s/api/v4/users/%s/image"
		AuthorIcon :=fmt.Sprintf(fmtstmnt, *siteURL, oldPost.UserId)
		if postUser.IsBot {
			botUser := model.BotFromUser(postUser)
			AuthorName = botUser.DisplayName
			AuthorIcon = fmt.Sprintf(fmtstmnt, *siteURL, botUser.UserId)
		}

		attachment := []*model.SlackAttachment{
			{
				Timestamp:  ateAt,
				AuthorName: AuthorName,
				AuthorIcon: AuthorIcon,
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
