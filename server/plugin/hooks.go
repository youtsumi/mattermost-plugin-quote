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

	if channel.Type == model.CHANNEL_DIRECT {
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

	matches := selfLinkPattern.FindAllString(post.Message, -1)
	if len(matches) != 0 {
		// Only first post matched the pattern is expanded, because can't deal with files that have more than five total attachments.
		match := matches[0]

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
		// NOTES: if attaching over 5 files, error will occur
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

		AuthorName := postUser.GetDisplayNameWithPrefix(model.SHOW_NICKNAME_FULLNAME, "@")
		fmtstmnt := "%s/api/v4/users/%s/image"
		AuthorIcon := fmt.Sprintf(fmtstmnt, *siteURL, oldPost.UserId)
		if postUser.IsBot {
			botUser := model.BotFromUser(postUser)
			AuthorName = botUser.DisplayName
			AuthorIcon = fmt.Sprintf(fmtstmnt, *siteURL, botUser.UserId)
		}

		attachment := []*model.SlackAttachment{
			{
				Timestamp:  oldPost.CreateAt,
				AuthorName: AuthorName,
				AuthorIcon: AuthorIcon,
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

	// Add additional comment written in dialog
	// If adding first the additional text in the message, the link in the additional text will be expanded, so additional text have to be added here
	if post.GetProp(postPropsKeyAdditionalText) != nil {
		post.Message = fmt.Sprintf("%s%s", post.GetProp(postPropsKeyAdditionalText), post.Message)
	}
	return post, ""
}
