# Share Post Plugin

This plugin enable to share and move Mattermost post to other channels.

## Usage
* 1. Click post dropdown menu and select `Share post` menu

![dropdown](./screenshots/dropdown.png)

* 2. Input dialog element and push `share` button
  * **Share to...**: The channel where selected post will be shared/moved
  * **Share type**:
    * **Share**: Share the post to selected channel
    * **Move**: Move post to selected channel, and delete original post
  * **Additionall Text**: Additional text for shared/moved post. Additional text will be inserted to a head of shared/moved post 

![dialog](./screenshots/dialog.png)

### Shared post
![shared_post](./screenshots/shared_post.png)

### Moved post
![moved_post](./screenshots/moved_post.png)


## Notes
* Creation time of moved post is the same as original post
* After sharing post, if original post is deleted, the link to original post is invalid
* Anyone can share/move posts created by others
  * The author of moved post will be the author of original post, (not user who move the post)
* If additional text has permalink to local post, it will be expanded and original post to share will not be expanded

## Limitation
* Only the first occurrence of the link will be expanded
* Cannot share/move the post to channeld in different team
* Cannot move the post that has parent/child posts (post thread)
  * Root post can be moved, but all children will be deleted
  * All children cannot be moved, because children must be the same channel as root post
* User cannot share/move the post to private channels / DM / GM
  * but the post in private channels / DM / GM can be shared/moved
* User can share/move the post to all public channel even though the user doesn't belong to the channel

## TODO
* Write tests
* **Might need to Mattermost changes**
  * Move thread
  * After sharing, redirect to the new post
  * Message attachments should be enalbe to render multi images
  * Footer link in MessageAttachments
