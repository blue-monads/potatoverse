# Engine API

## Engine

### GET /zz/api/core/engine/debug

Get engine debug data.

**Response:** Debug data object

### GET /zz/api/core/engine/space_info/:space_key

Get space info.

**Query:**
- `host_name` (string) - Host name

**Response:** Space info object

### GET /zz/api/core/engine/derivehost/:nskey

Derive host.

**Query:**
- `host_name` (string) - Host name
- `space_id` (int64) - Space ID

**Response:**
- `host` (string) - Derived host
- `space_id` (int64) - Space ID

## Repo

### GET /zz/api/core/repo/list

List repos.

**Response:** Array of repos

## File Upload

### POST /zz/file/upload-presigned

Upload file with presigned token.

**Query:**
- `presigned-key` (string) - Presigned token

**Request:** Multipart form with `file`

**Response:**
- `file_id` (int64) - File ID
- `message` (string) - Success message

## Extra Routes

### GET /zz/profileImage/:id/:name

Get user profile image SVG.

**Response:** SVG image

### GET /zz/profileImage/:id

Get user profile image SVG by ID.

**Response:** SVG image

### GET /zz/api/gradients

List gradients.

**Response:** Array of gradient objects



