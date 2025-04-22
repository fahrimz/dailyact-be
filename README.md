# Daily Activities Backend

A Go-based backend for a daily note-taking application with activity tracking and categorization.

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (recommended)
- PostgreSQL database (if not using Docker)

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/fahrimz/dailyact-be.git
   cd dailyact-be
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env
   ```
   Modify the `.env` file with your desired configuration.

3. Start the database (choose one):

   a. Using Docker (recommended):
   ```bash
   docker-compose up -d
   ```

   b. Without Docker:
   Create a PostgreSQL database and update the `.env` file with your database credentials.

## Running the Application

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Run the application:
   ```bash
   go run main.go
   ```

The server will start on port 8080.

## API Endpoints

### Categories
- `POST /categories` - Create a new category
- `GET /categories` - List all categories

### Activities
- `POST /activities` - Create a new activity
- `GET /activities` - List all activities
- `GET /activities/:id` - Get a specific activity
- `PUT /activities/:id` - Update an activity
- `DELETE /activities/:id` - Delete an activity

## Data Models

### Category
```json
{
  "name": "Work",
  "description": "Work-related activities"
}
```

### Activity
```json
{
  "date": "2025-04-22T00:00:00Z",
  "start_time": "2025-04-22T09:00:00Z",
  "end_time": "2025-04-22T10:30:00Z",
  "duration": 90,
  "description": "Team meeting",
  "notes": "Discussed project timeline",
  "category_id": 1
}
```
