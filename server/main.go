package main

import (
	"github.com/kaakaa/mattermost-plugin-share-post/server/plugin"
	mmplugin "github.com/mattermost/mattermost-server/v5/plugin"
)

func main() {
	mmplugin.ClientMain(&plugin.SharePostPlugin{})
}
