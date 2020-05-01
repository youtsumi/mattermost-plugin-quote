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
	p.API.LogWarn("XXXX")
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

		postUser, err := p.API.GetUser(oldPost.UserId)
		if err != nil {
			return post, err.Message
		}

		quote := fmt.Sprintf("**%s** at **%s** said:\n", postUser.Nickname, time.Unix(oldPost.CreateAt, 0))
		messageLines := strings.Split(oldPost.Message, "\n")
		for _, line := range messageLines {
			quote = fmt.Sprintf("%s\n> %s", quote, line)
		}
		post.Message = strings.Replace(post.Message, match, quote, 1)
	}

	return post, ""

}
