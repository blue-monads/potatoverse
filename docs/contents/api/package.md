# Package API

## Package

### POST /zz/api/core/package/install

Install package from URL.

**Request:**
- `url` (string) - Package URL

**Response:**
- `installed_id` (int64) - Installed package ID
- `root_space_id` (int64) - Root space ID
- `key_space` (string) - Key space
- `init_page` (string) - Init page

### POST /zz/api/core/package/install/zip

Install package from ZIP file.

**Request:** ZIP file in body

**Response:** InstallPackageResult object

### POST /zz/api/core/package/install/embed

Install embedded package.

**Request:**
- `name` (string) - Package name
- `repo_slug` (string) - Repo slug (optional)

**Response:** InstallPackageResult object

### DELETE /zz/api/core/package/:id

Delete package.

**Response:** null

### POST /zz/api/core/package/:id/dev-token

Generate package dev token.

**Response:**
- `token` (string) - Dev token

### POST /zz/api/core/package/push

Push package update (requires dev token in Authorization header).

**Query:**
- `recreate_artifacts` (bool) - Recreate artifacts

**Request:** ZIP file in body

**Response:**
- `package_version_id` (int64) - Package version ID
- `message` (string) - Success message

### GET /zz/api/core/package/list

List packages.

**Query:**
- `repo` (string) - Repo slug filter

**Response:** Array of packages

### GET /zz/api/core/package/:id/info

Get installed package info.

**Response:** Package info object

## Package Files

### GET /zz/api/core/vpackage/:id/files

List package files.

**Query:**
- `path` (string) - Path filter

**Response:** Array of files

### GET /zz/api/core/vpackage/:id/files/:fileId

Get package file.

**Response:** File object

### GET /zz/api/core/vpackage/:id/files/:fileId/download

Download package file.

**Response:** File content

### DELETE /zz/api/core/vpackage/:id/files/:fileId

Delete package file.

**Response:** Success message

### POST /zz/api/core/vpackage/:id/files/upload

Upload package file.

**Request:** Multipart form
- `file` - File to upload
- `path` (string) - Target path

**Response:**
- `file_id` (int64) - File ID
- `message` (string) - Success message



