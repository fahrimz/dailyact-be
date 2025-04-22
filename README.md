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

## Authentication

This API uses Google OAuth2 for authentication.

### Complete Login Flow with Frontend

Here's how the authentication works in a complete frontend + backend setup:

#### Frontend Implementation Options

1. **Browser Redirect (Recommended for Web & Mobile)**
   ```javascript
   const handleGoogleLogin = async () => {
     // 1. Get login URL from backend
     const response = await axios.get('/auth/google/login');
     const { url } = response.data.data;

     // 2. Redirect to Google login
     window.location.href = url;
     // For mobile apps: use in-app browser or system browser
     // iOS: ASWebAuthenticationSession
     // Android: Chrome Custom Tabs
   };
   ```

2. **Popup Window (Web Only)**
   ```javascript
   const handleGoogleLoginPopup = async () => {
     const response = await axios.get('/auth/google/login');
     const { url } = response.data.data;

     window.open(url, 'Google Login', 'width=500,height=600');
   };
   ```

3. **Mobile SDK Integration**
   - iOS: Use Google Sign-In SDK
   - Android: Use Google Sign-In SDK
   - Send obtained ID token to backend for verification:
     ```http
     POST /auth/google/verify
     Content-Type: application/json

     {
       "id_token": "eyJhbGciOiJSUzI1..."
     }

     Response:
     {
       "success": true,
       "message": "Login successful",
       "data": {
         "token": "your.jwt.token",
         "user": {
           "id": 1,
           "email": "user@example.com",
           "name": "John Doe",
           "picture": "https://...",
           "role": "user"
         }
       }
     }
     ```

#### Backend Flow

1. **Initial Login Request**
   - Frontend application sends GET request to `/auth/google/login`
   - Backend API generates a Google OAuth URL with required scopes (email, profile)
   - Backend API returns the URL to the frontend

2. **Google Authentication**
   - Browser redirects to the Google login page
   - User enters their Google credentials
   - Google validates the credentials
   - If valid, Google asks user to consent to sharing their information

3. **OAuth Callback**
   - After user consent, Google redirects browser back to `/auth/google/callback`
   - Authorization code is included in the redirect URL
   - Backend API exchanges this code with Google for an access token
   - Backend API uses the access token to fetch user's information (email, name, picture)

4. **User Creation/Update**
   - Backend API checks if a user with the Google ID exists in database
   - If not found: creates new user with 'user' role
   - If found: updates last login timestamp
   - Stores user's email, name, and profile picture

5. **JWT Generation**
   - Backend API generates a JWT token containing:
     - User ID
     - Email
     - Role (user/admin)
     - Expiration time (24 hours)
   - Token is signed with backend's secret key

6. **Response & Frontend Integration**
   - Backend returns JWT token and user information
   - Frontend stores token in localStorage/secure cookie
   - Frontend updates UI to show user is logged in
   - Frontend includes token in all subsequent API requests:
     ```javascript
     // Example using axios
     axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
     ```

#### Implementation Examples

1. **Web Implementation (React)**
```javascript
// Using redirect approach
const handleGoogleLogin = async () => {
  try {
    // Store the return URL for after login
    localStorage.setItem('returnTo', window.location.pathname);
    
    // Get login URL and redirect
    const response = await axios.get('/auth/google/login');
    window.location.href = response.data.data.url;
  } catch (error) {
    console.error('Login failed:', error);
  }
};

// Handle callback in your callback route component
const GoogleCallback = () => {
  useEffect(() => {
    const handleCallback = async () => {
      // Get token from URL params (handled by backend)
      const { token, user } = await getCurrentUser();
      
      // Store authentication state
      setAuthState({ token, user });
      
      // Redirect back to original page
      const returnTo = localStorage.getItem('returnTo') || '/dashboard';
      navigate(returnTo);
    };
    
    handleCallback();
  }, []);
  
  return <LoadingSpinner />;
};
```

2. **Mobile Implementation (React Native with Google Sign-In)**
```javascript
import { GoogleSignin } from '@react-native-google-signin/google-signin';

// Initialize Google Sign-In in your App.js or similar
GoogleSignin.configure({
  webClientId: 'YOUR_WEB_CLIENT_ID.apps.googleusercontent.com',
  iosClientId: 'YOUR_IOS_CLIENT_ID.apps.googleusercontent.com',
});

const handleGoogleLogin = async () => {
  try {
    // Start Google Sign-In flow
    await GoogleSignin.hasPlayServices();
    await GoogleSignin.signIn();
    
    // Get ID token
    const { idToken } = await GoogleSignin.getTokens();
    
    // Verify token with backend
    const response = await axios.post('/auth/google/verify', {
      id_token: idToken
    });
    
    const { token, user } = response.data.data;
    
    // Store JWT securely
    await SecureStore.setItemAsync('token', token);
    
    // Set default auth header
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    
    // Update app state
    setUser(user);
    navigation.replace('Home');
    
  } catch (error) {
    if (error.code === statusCodes.SIGN_IN_CANCELLED) {
      console.log('User cancelled login flow');
    } else if (error.code === statusCodes.IN_PROGRESS) {
      console.log('Sign in is in progress');
    } else {
      console.error('Login failed:', error);
    }
  }
};

// Logout function
const handleLogout = async () => {
  try {
    // Sign out from Google
    await GoogleSignin.signOut();
    
    // Clear stored token
    await SecureStore.deleteItemAsync('token');
    
    // Clear auth header
    delete axios.defaults.headers.common['Authorization'];
    
    // Update app state
    setUser(null);
    navigation.replace('Login');
  } catch (error) {
    console.error('Logout failed:', error);
  }
};
```

#### Security Considerations
1. **Token Storage**
   - Store JWT in httpOnly cookies for better security
   - Clear token on logout/expiration

2. **CORS Configuration**
   - Backend must allow requests from frontend domain
   - Properly configure allowed origins

3. **Error Handling**
   - Handle token expiration gracefully
   - Implement refresh token mechanism if needed
   - Show appropriate error messages to users

### Authentication Flow Example

1. Get the login URL:
   ```http
   GET /auth/google/login

   Response:
   {
     "success": true,
     "message": "Login URL generated",
     "data": {
       "url": "https://accounts.google.com/o/oauth2/v2/auth?...."
     }
   }
   ```

2. After following the URL and logging in with Google, you'll be redirected to the callback URL with a success response:
   ```http
   GET /auth/google/callback?code=.....

   Response:
   {
     "success": true,
     "message": "Login successful",
     "data": {
       "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9....",
       "user": {
         "id": 1,
         "email": "user@example.com",
         "name": "John Doe",
         "picture": "https://....",
         "role": "user",
         "created_at": "2025-04-22T02:11:45Z",
         "last_login_at": "2025-04-22T02:11:45Z"
       }
     }
   }
   ```

3. Use the JWT token in subsequent requests:
   ```http
   GET /activities
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9....
   ```

### User Roles
- **User**: Can manage their own activities
- **Admin**: Can manage all activities and categories

## API Endpoints

### Auth
- `GET /auth/google/login` - Get Google login URL
- `GET /auth/google/callback` - Google OAuth callback
- `GET /auth/me` - Get current user info 🔒

### Users
- `GET /users` - List all users 🔒👑
  - Query parameters:
    - `page` (optional, default: 1) - Page number
    - `page_size` (optional, default: 10, max: 100) - Number of items per page
- `GET /users/:id` - Get user details with their activities 🔒👑

### Categories
- `POST /categories` - Create a new category
- `GET /categories` - List all categories
  - Query parameters:
    - `page` (optional, default: 1) - Page number
    - `page_size` (optional, default: 10, max: 100) - Number of items per page

### Activities
- `POST /activities` - Create a new activity 🔒
- `GET /activities` - List activities 🔒
  - For admin: Lists all activities
  - For users: Lists only their activities
  - Query parameters:
    - `page` (optional, default: 1) - Page number
    - `page_size` (optional, default: 10, max: 100) - Number of items per page
- `GET /activities/:id` - Get a specific activity 🔒👤
- `PUT /activities/:id` - Update an activity 🔒👤
- `DELETE /activities/:id` - Delete an activity 🔒👤

Legend:
- 🔒 Requires authentication
- 👑 Requires admin role
- 👤 Requires ownership or admin role

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation successful message",
  "data": { },
  "pagination": {
    "current_page": 1,
    "page_size": 10,
    "total_items": 50,
    "total_pages": 5,
    "has_more": true
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message",
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "detail": "Detailed error information"
  }
}
```

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
