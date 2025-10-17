# TaskFolio API Specification

## Overview
This document defines the RESTful API endpoints for TaskFolio.

## Base URL
```
Development: http://localhost:8080/api/v1
Production: https://api.taskfolio.com/v1
```

## Authentication
All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

---

## Authentication Endpoints

### POST /auth/register
Register a new user.

**Request Body:**
```json
{
  "email": "student@example.com",
  "username": "johndoe",
  "password": "SecurePassword123!",
  "firstName": "John",
  "lastName": "Doe",
  "role": "student"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "email": "student@example.com",
  "username": "johndoe",
  "firstName": "John",
  "lastName": "Doe",
  "role": "student",
  "token": "jwt-token"
}
```

### POST /auth/login
Authenticate user and receive token.

**Request Body:**
```json
{
  "email": "student@example.com",
  "password": "SecurePassword123!"
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": "uuid",
    "email": "student@example.com",
    "username": "johndoe",
    "role": "student"
  },
  "token": "jwt-token",
  "expiresAt": "2025-10-17T18:54:52Z"
}
```

### POST /auth/logout
Logout user (invalidate token).

**Response:** `200 OK`

### POST /auth/refresh
Refresh JWT token.

**Response:** `200 OK`
```json
{
  "token": "new-jwt-token",
  "expiresAt": "2025-10-17T18:54:52Z"
}
```

### POST /auth/forgot-password
Request password reset email.

**Request Body:**
```json
{
  "email": "student@example.com"
}
```

**Response:** `200 OK`

### POST /auth/reset-password
Reset password with token.

**Request Body:**
```json
{
  "token": "reset-token",
  "newPassword": "NewSecurePassword123!"
}
```

**Response:** `200 OK`

---

## User Endpoints

### GET /users/me
Get current user profile.

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "student@example.com",
  "username": "johndoe",
  "firstName": "John",
  "lastName": "Doe",
  "role": "student",
  "avatarUrl": "https://...",
  "createdAt": "2025-01-01T00:00:00Z"
}
```

### PUT /users/me
Update current user profile.

**Request Body:**
```json
{
  "firstName": "John",
  "lastName": "Doe",
  "avatarUrl": "https://..."
}
```

**Response:** `200 OK`

### GET /users/:id
Get user by ID (admin/instructor only).

**Response:** `200 OK`

### GET /users
List users with pagination and filters.

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `role`: Filter by role
- `search`: Search by name or email

**Response:** `200 OK`
```json
{
  "users": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

---

## Task Endpoints

### POST /tasks
Create a new task.

**Request Body:**
```json
{
  "title": "Complete assignment",
  "description": "Finish the Go programming assignment",
  "status": "todo",
  "priority": "high",
  "assignedTo": "user-uuid",
  "dueDate": "2025-10-20T23:59:59Z",
  "tags": ["uuid1", "uuid2"],
  "estimatedHours": 5.0
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "title": "Complete assignment",
  "status": "todo",
  "priority": "high",
  "createdBy": "uuid",
  "assignedTo": "uuid",
  "dueDate": "2025-10-20T23:59:59Z",
  "createdAt": "2025-10-16T18:54:52Z"
}
```

### GET /tasks
List tasks with filters and pagination.

**Query Parameters:**
- `page`: Page number
- `limit`: Items per page
- `status`: Filter by status
- `priority`: Filter by priority
- `assignedTo`: Filter by assigned user
- `createdBy`: Filter by creator
- `tags`: Filter by tags (comma-separated UUIDs)
- `search`: Search in title and description
- `dueFrom`: Filter tasks due after date
- `dueTo`: Filter tasks due before date
- `sort`: Sort field (dueDate, priority, createdAt, title)
- `order`: Sort order (asc, desc)

**Response:** `200 OK`
```json
{
  "tasks": [
    {
      "id": "uuid",
      "title": "Complete assignment",
      "status": "todo",
      "priority": "high",
      "dueDate": "2025-10-20T23:59:59Z",
      "createdBy": {...},
      "assignedTo": {...},
      "tags": [...]
    }
  ],
  "pagination": {...}
}
```

### GET /tasks/:id
Get task details by ID.

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "title": "Complete assignment",
  "description": "Finish the Go programming assignment",
  "status": "todo",
  "priority": "high",
  "createdBy": {...},
  "assignedTo": {...},
  "tags": [...],
  "comments": [...],
  "attachments": [...],
  "subtasks": [...],
  "history": [...]
}
```

### PUT /tasks/:id
Update a task.

**Request Body:**
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "in_progress",
  "priority": "high"
}
```

**Response:** `200 OK`

### DELETE /tasks/:id
Delete a task.

**Response:** `204 No Content`

### POST /tasks/:id/assign
Assign task to a user.

**Request Body:**
```json
{
  "userId": "uuid"
}
```

**Response:** `200 OK`

### POST /tasks/:id/complete
Mark task as completed.

**Response:** `200 OK`

### POST /tasks/:id/duplicate
Duplicate a task.

**Response:** `201 Created`

---

## Comment Endpoints

### POST /tasks/:taskId/comments
Add a comment to a task.

**Request Body:**
```json
{
  "content": "This is a comment",
  "parentCommentId": "uuid" // optional, for threaded comments
}
```

**Response:** `201 Created`

### GET /tasks/:taskId/comments
Get comments for a task.

**Response:** `200 OK`
```json
{
  "comments": [
    {
      "id": "uuid",
      "content": "This is a comment",
      "user": {...},
      "createdAt": "2025-10-16T18:54:52Z",
      "replies": [...]
    }
  ]
}
```

### PUT /comments/:id
Update a comment.

**Response:** `200 OK`

### DELETE /comments/:id
Delete a comment.

**Response:** `204 No Content`

---

## Tag Endpoints

### POST /tags
Create a new tag.

**Request Body:**
```json
{
  "name": "Frontend",
  "color": "#3B82F6"
}
```

**Response:** `201 Created`

### GET /tags
List all tags.

**Response:** `200 OK`
```json
{
  "tags": [
    {
      "id": "uuid",
      "name": "Frontend",
      "color": "#3B82F6"
    }
  ]
}
```

### PUT /tags/:id
Update a tag.

**Response:** `200 OK`

### DELETE /tags/:id
Delete a tag.

**Response:** `204 No Content`

---

## Notification Endpoints

### GET /notifications
Get user's notifications.

**Query Parameters:**
- `page`: Page number
- `limit`: Items per page
- `isRead`: Filter by read status

**Response:** `200 OK`
```json
{
  "notifications": [
    {
      "id": "uuid",
      "type": "task_assigned",
      "title": "New task assigned",
      "message": "You have been assigned a new task",
      "link": "/tasks/uuid",
      "isRead": false,
      "createdAt": "2025-10-16T18:54:52Z"
    }
  ],
  "unreadCount": 5
}
```

### PUT /notifications/:id/read
Mark notification as read.

**Response:** `200 OK`

### PUT /notifications/read-all
Mark all notifications as read.

**Response:** `200 OK`

---

## Team Endpoints

### POST /teams
Create a new team.

**Request Body:**
```json
{
  "name": "Frontend Team",
  "description": "Team responsible for frontend development"
}
```

**Response:** `201 Created`

### GET /teams
List teams.

**Response:** `200 OK`

### GET /teams/:id
Get team details.

**Response:** `200 OK`

### PUT /teams/:id
Update team.

**Response:** `200 OK`

### DELETE /teams/:id
Delete team.

**Response:** `204 No Content`

### POST /teams/:id/members
Add member to team.

**Request Body:**
```json
{
  "userId": "uuid",
  "role": "member"
}
```

**Response:** `200 OK`

### DELETE /teams/:id/members/:userId
Remove member from team.

**Response:** `204 No Content`

---

## Analytics Endpoints

### GET /analytics/dashboard
Get dashboard analytics.

**Response:** `200 OK`
```json
{
  "totalTasks": 150,
  "completedTasks": 75,
  "inProgressTasks": 50,
  "overdueTasks": 10,
  "completionRate": 50.0,
  "avgCompletionTime": 48.5
}
```

### GET /analytics/tasks/by-status
Get task count by status.

**Response:** `200 OK`

### GET /analytics/tasks/by-priority
Get task count by priority.

**Response:** `200 OK`

### GET /analytics/productivity
Get user productivity metrics.

**Query Parameters:**
- `userId`: User ID
- `from`: Start date
- `to`: End date

**Response:** `200 OK`

---

## Attachment Endpoints

### POST /tasks/:taskId/attachments
Upload file attachment.

**Request:** Multipart form data with file

**Response:** `201 Created`

### GET /attachments/:id
Download attachment.

**Response:** File download

### DELETE /attachments/:id
Delete attachment.

**Response:** `204 No Content`

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "field": "email",
      "issue": "Invalid email format"
    }
  }
}
```

### 401 Unauthorized
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

### 403 Forbidden
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "You don't have permission to access this resource"
  }
}
```

### 404 Not Found
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

### 500 Internal Server Error
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred"
  }
}
```

---

## Rate Limiting

- **Public endpoints**: 100 requests per 15 minutes
- **Authenticated endpoints**: 1000 requests per 15 minutes
- **File upload endpoints**: 20 requests per 15 minutes

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1634400000
```

---

## Pagination

All list endpoints support pagination with the following parameters:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

Response includes pagination metadata:
```json
{
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5,
    "hasNext": true,
    "hasPrev": false
  }
}
```

---

## WebSocket API

### Connection
```
ws://localhost:8080/ws?token=<jwt-token>
```

### Events

**Client to Server:**
- `subscribe:task:<taskId>` - Subscribe to task updates
- `subscribe:notifications` - Subscribe to notifications
- `ping` - Keep connection alive

**Server to Client:**
- `task:updated` - Task was updated
- `task:assigned` - Task was assigned
- `notification:new` - New notification
- `pong` - Response to ping
