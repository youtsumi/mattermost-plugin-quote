import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getCurrentChannel} from 'mattermost-redux/selectors/entities/channels'
import {isOpenChannel} from 'mattermost-redux/utils/channel_utils'

import {id as pluginId} from './manifest';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        registry.registerPostDropdownMenuAction(
            'Share post',
            (postId) => {
                const extra_elements = [];
                if (!isOpenChannel(getCurrentChannel(store.getState()))) {
                    extra_elements.push({
                        display_name: "This channel is not public. Are you sure to share this post to other channel?",
                        name: 'force_share',
                        type: 'bool',
                        placeholder: 'Yes, I confirm that this post share to other channel.',
                    });
                }
                window.openInteractiveDialog({
                    url: getPluginServerRoute(store.getState()) + '/api/v1/share',
                    dialog: {
                        callback_id: postId,
                        title: 'Share post',
                        elements: [{
                            display_name: 'Share to...',
                            name: 'to_channel',
                            type: 'select',
                            data_source: 'channels',
                            placeholder: 'Find a channel to share',
                        }, ...extra_elements,
                        {
                            display_name: 'Share type',
                            name: 'share_type',
                            type: 'radio',
                            default: 'share',
                            options: [{
                                text: 'Share',
                                value: 'share',
                            }, {
                                text: 'Move',
                                value: 'move',
                            }]
                        },{
                            display_name: 'Additional Text',
                            name: 'additional_text',
                            type: 'textarea',
                            optional: true,
                            placeholder: 'Write an additional text (optional)',
                        }],
                        submit_label: 'Share',
                    }
                });
            }
        );
    }
}

const getPluginServerRoute = (state) => {
    const config = getConfig(state);

    let basePath = '/';
    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return basePath + '/plugins/' + pluginId;
};

window.registerPlugin(pluginId, new Plugin());
