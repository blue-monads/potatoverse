# Space API

## Space

### GET /zz/api/core/space/installed

List installed spaces.

**Response:** Array of spaces

### POST /zz/api/core/space/authorize/:space_key

Authorize space access.

**Request:**
- `package_id` (int64) - Package ID
- `space_id` (int64) - Space ID

**Response:**
- `token` (string) - Space token

## Space KV

### GET /zz/api/core/space/:install_id/kv

List space KV entries.

**Query:**
- `group` (string) - Group filter
- `key` (string) - Key filter
- `tag1` (string) - Tag1 filter
- `tag2` (string) - Tag2 filter
- `tag3` (string) - Tag3 filter
- `offset` (int) - Offset
- `limit` (int) - Limit (default: 100)

**Response:** Array of KV entries

### GET /zz/api/core/space/:install_id/kv/:kvId

Get KV entry.

**Response:** KV entry object

### POST /zz/api/core/space/:install_id/kv

Create KV entry.

**Request:** Map of KV data

**Response:** KV entry object

### PUT /zz/api/core/space/:install_id/kv/:kvId

Update KV entry.

**Request:** Map of fields to update

**Response:** KV entry object

### DELETE /zz/api/core/space/:install_id/kv/:kvId

Delete KV entry.

**Response:** Success message

## Space Files

### GET /zz/api/core/space/:install_id/files

List space files.

**Query:**
- `path` (string) - Path filter

**Response:** Array of files

### GET /zz/api/core/space/:install_id/files/:fileId

Get space file.

**Response:** File object

### GET /zz/api/core/space/:install_id/files/:fileId/download

Download space file.

**Response:** File content

### DELETE /zz/api/core/space/:install_id/files/:fileId

Delete space file.

**Response:** Success message

### POST /zz/api/core/space/:install_id/files/upload

Upload space file.

**Request:** Multipart form
- `file` - File to upload
- `path` (string) - Target path

**Response:**
- `file_id` (int64) - File ID
- `message` (string) - Success message

### POST /zz/api/core/space/:install_id/files/folder

Create folder.

**Request:**
- `name` (string) - Folder name
- `path` (string) - Parent path

**Response:**
- `folder_id` (int64) - Folder ID
- `message` (string) - Success message

### POST /zz/api/core/space/:install_id/files/presigned

Create presigned upload URL.

**Request:**
- `file_name` (string) - File name
- `path` (string) - Target path
- `expiry` (int64) - Expiry in seconds (default: 3600)

**Response:**
- `presigned_token` (string) - Presigned token
- `upload_url` (string) - Upload URL
- `expiry` (int64) - Expiry time

## Space Users

### GET /zz/api/core/space/:install_id/users

List space users.

**Query:**
- `space_id` (int64) - Space ID filter
- `user_id` (int64) - User ID filter
- `scope` (string) - Scope filter

**Response:** Array of space users

### GET /zz/api/core/space/:install_id/users/:spaceUserId

Get space user.

**Response:** Space user object

### POST /zz/api/core/space/:install_id/users

Create space user.

**Request:** Map of user data

**Response:** Space user object

### PUT /zz/api/core/space/:install_id/users/:spaceUserId

Update space user.

**Request:** Map of fields to update

**Response:** Space user object

### DELETE /zz/api/core/space/:install_id/users/:spaceUserId

Delete space user.

**Response:** Success message

## Space Capabilities

### GET /zz/api/core/space/:install_id/capabilities

List space capabilities.

**Query:**
- `capability_type` (string) - Capability type filter
- `space_id` (int64) - Space ID filter (0 for package-level)

**Response:** Array of capabilities

### GET /zz/api/core/space/:install_id/capabilities/:capabilityId

Get capability.

**Response:** Capability object

### POST /zz/api/core/space/:install_id/capabilities

Create capability.

**Request:** Map of capability data (space_id defaults to 0 if omitted)

**Response:** Capability object

### PUT /zz/api/core/space/:install_id/capabilities/:capabilityId

Update capability.

**Request:** Map of fields to update

**Response:** Capability object

### DELETE /zz/api/core/space/:install_id/capabilities/:capabilityId

Delete capability.

**Response:** Success message

## Space Events

### GET /zz/api/core/space/:install_id/events

List event subscriptions.

**Query:**
- `space_id` (int64) - Space ID filter
- `event_key` (string) - Event key filter

**Response:** Array of subscriptions

### GET /zz/api/core/space/:install_id/events/:subscriptionId

Get event subscription.

**Response:** Subscription object

### POST /zz/api/core/space/:install_id/events

Create event subscription.

**Request:** Subscription object

**Response:** Subscription object

### PUT /zz/api/core/space/:install_id/events/:subscriptionId

Update event subscription.

**Request:** Map of fields to update

**Response:** Subscription object

### DELETE /zz/api/core/space/:install_id/events/:subscriptionId

Delete event subscription.

**Response:** Success message

