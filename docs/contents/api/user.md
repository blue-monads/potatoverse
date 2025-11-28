# User API

## User

### GET /zz/api/core/user

List users.

**Query:**
- `offset` (int) - Offset
- `limit` (int) - Limit (default: 100)

**Response:** Array of users

### GET /zz/api/core/user/:id

Get user by ID.

**Response:** User object

### POST /zz/api/core/user/create

Create user directly.

**Request:**
- `name` (string) - User name
- `email` (string) - Email
- `username` (string) - Username
- `utype` (string) - User type
- `ugroup` (string) - User group

**Response:** User object

## User Invites

### GET /zz/api/core/user/invites

List user invites.

**Query:**
- `offset` (int) - Offset
- `limit` (int) - Limit (default: 100)

**Response:** Array of invites

### GET /zz/api/core/user/invites/:id

Get invite by ID.

**Response:** Invite object

### POST /zz/api/core/user/invites

Create invite.

**Request:**
- `email` (string) - Email
- `invited_as_type` (string) - Invite type

**Response:** Invite object

### PUT /zz/api/core/user/invites/:id

Update invite.

**Request:** Map of fields to update

**Response:** Success message

### DELETE /zz/api/core/user/invites/:id

Delete invite.

**Response:** Success message

### POST /zz/api/core/user/invites/:id/resend

Resend invite.

**Response:** Invite object

## User Groups

### GET /zz/api/core/user/groups

List user groups.

**Response:** Array of groups

### GET /zz/api/core/user/groups/:name

Get user group by name.

**Response:** Group object

### POST /zz/api/core/user/groups

Create user group.

**Request:**
- `name` (string) - Group name
- `info` (string) - Group info

**Response:** Success message

### PUT /zz/api/core/user/groups/:name

Update user group.

**Request:**
- `info` (string) - Group info

**Response:** Success message

### DELETE /zz/api/core/user/groups/:name

Delete user group.

**Response:** Success message

## User Messages

### GET /zz/api/core/user/messages

List user messages.

**Query:**
- `after_id` (int64) - Message ID to start after
- `limit` (int) - Limit (default: 100)

**Response:** Array of messages

### GET /zz/api/core/user/messages/new

Query new messages.

**Response:** Array of new messages

### GET /zz/api/core/user/messages/history

Query message history.

**Query:**
- `limit` (int) - Limit (default: 100, max: 1000)

**Response:** Array of messages

### POST /zz/api/core/user/messages/read-all

Mark all messages as read.

**Response:** Success message with read_head

### POST /zz/api/core/user/messages

Send message.

**Request:**
- `title` (string) - Message title
- `type` (string) - Message type
- `contents` (string) - Message contents
- `to_user` (int64) - Recipient user ID
- `from_space_id` (int64) - Source space ID (optional)
- `callback_token` (string) - Callback token (optional)
- `warn_level` (int) - Warning level (optional)

**Response:** Message object

### GET /zz/api/core/user/messages/:id

Get message by ID.

**Response:** Message object

### PUT /zz/api/core/user/messages/:id

Update message.

**Request:** Map of fields to update

**Response:** Success message

### DELETE /zz/api/core/user/messages/:id

Delete message.

**Response:** Success message

### POST /zz/api/core/user/messages/:id/read

Mark message as read.

**Response:** Success message

## Self User

### GET /zz/api/core/self/info

Get current user info.

**Response:** User object

### GET /zz/api/core/self/portalData/:portal_type

Get portal data.

**Response:**
- `popular_keywords` (array) - Popular keywords
- `favorite_projects` (array) - Favorite projects

### PUT /zz/api/core/self/bio

Update bio.

**Request:**
- `bio` (string) - Bio text

**Response:** Success message



