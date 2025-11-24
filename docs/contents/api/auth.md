# Auth API

### POST /zz/api/core/auth/login

Login with username/email and password.

**Request:**
- `username` (string) - Username or email
- `password` (string) - Password

**Response:**
- `access_token` (string) - Access token
- `user_info` (object) - User information
- `portal_page_type` (string) - Portal page type

### GET /zz/api/core/auth/invite/:token

Get invite information.

**Response:**
- `email` (string) - Invite email
- `role` (string) - Invited role
- `expires_on` (string) - Expiration date

### POST /zz/api/core/auth/invite/:token

Accept invite and create account.

**Request:**
- `name` (string) - User name
- `username` (string) - Username
- `password` (string) - Password (min 6 chars)

**Response:**
- `message` (string) - Success message
- `user` (object) - Created user



