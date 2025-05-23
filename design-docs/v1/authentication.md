# Authentication & Authorization

**Related Documents:**
- [API Design](./api-design.md) - How auth integrates with API
- [Database Schema](./database-schema.md) - User and auth-related tables
- [Architecture](./architecture.md) - System architecture context

## Authentication Strategy

### OAuth Strategy
- **Primary Provider**: Google OAuth 2.0
- **Development Mode**: Dev/Test login allowing email input
- **Extensible Design**: Interface to support multiple providers
- **Key Requirements**: Extract verified email address

### Session Management (Simple JWT for MVP)
- **Access Tokens**: JWT with 7-day expiry (no refresh tokens initially)
- **Storage**: Frontend localStorage with fallback to sessionStorage
- **Concurrent Sessions**: Supported across multiple devices
- **Token Claims**: User ID, email, expiry timestamp
- **Migration Path**: Adding refresh tokens later requires minimal changes

## JWT Token Structure

```go
// JWT Token Structure
type JWTClaims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    Provider  string `json:"provider"`
    ExpiresAt int64  `json:"exp"`
    IssuedAt  int64  `json:"iat"`
    jwt.StandardClaims
}

// JWT Configuration
type JWTConfig struct {
    SecretKey    string        // From environment
    ExpiryTime   time.Duration // 7 days for MVP
    Issuer       string        // "tribe-app"
}
```

## OAuth Implementation

### Google OAuth Flow

```go
// OAuth Configuration
type OAuthConfig struct {
    GoogleClientID     string
    GoogleClientSecret string
    RedirectURL        string
    Scopes            []string // ["openid", "email", "profile"]
}

// OAuth Handler
func (a *AuthService) HandleGoogleOAuth(c *gin.Context) {
    // 1. Validate OAuth state parameter
    state := c.Query("state")
    if !a.validateState(state) {
        c.JSON(400, gin.H{"error": "Invalid state parameter"})
        return
    }
    
    // 2. Exchange code for tokens
    code := c.Query("code")
    token, err := a.oauth.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to exchange code"})
        return
    }
    
    // 3. Get user info from Google
    userInfo, err := a.getUserInfo(token.AccessToken)
    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to get user info"})
        return
    }
    
    // 4. Create or update user in database
    user, err := a.createOrUpdateUser(userInfo)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create user"})
        return
    }
    
    // 5. Generate JWT token
    jwtToken, err := a.generateJWT(user)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to generate token"})
        return
    }
    
    // 6. Return token to frontend
    c.JSON(200, gin.H{
        "token": jwtToken,
        "user":  user,
    })
}
```

### Development OAuth (For Testing)

```go
// Development login for testing
func (a *AuthService) HandleDevLogin(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required,email"`
        Name  string `json:"name" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Create development user
    user := &User{
        Email:     req.Email,
        Name:      req.Name,
        Provider:  "dev",
        OAuthID:   req.Email, // Use email as OAuth ID for dev
        Verified:  true,      // Auto-verify dev users
    }
    
    user, err := a.createOrUpdateUser(user)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create dev user"})
        return
    }
    
    jwtToken, err := a.generateJWT(user)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to generate token"})
        return
    }
    
    c.JSON(200, gin.H{
        "token": jwtToken,
        "user":  user,
    })
}
```

## JWT Implementation

### Token Generation

```go
func (a *AuthService) generateJWT(user *User) (string, error) {
    claims := JWTClaims{
        UserID:   user.ID,
        Email:    user.Email,
        Provider: user.Provider,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(a.config.ExpiryTime).Unix(),
            IssuedAt:  time.Now().Unix(),
            Issuer:    a.config.Issuer,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(a.config.SecretKey))
}
```

### Token Validation

```go
func (a *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(a.config.SecretKey), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

### Middleware Implementation

```go
func (a *AuthService) JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // Extract Bearer token
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }
        
        claims, err := a.ValidateJWT(tokenParts[1])
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Add user info to context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_provider", claims.Provider)
        
        c.Next()
    }
}
```

## Authorization Model

### Resource-Based Authorization

```go
// Authorization service
type AuthorizationService struct {
    db repository.Database
}

// Check if user owns or has access to a list
func (az *AuthorizationService) CanAccessList(userID, listID string) (bool, error) {
    list, err := az.db.GetList(listID)
    if err != nil {
        return false, err
    }
    
    // Check direct ownership
    if list.OwnerType == "user" && list.OwnerID == userID {
        return true, nil
    }
    
    // Check tribe membership for tribe lists
    if list.OwnerType == "tribe" {
        isMember, err := az.db.IsUserTribeMember(userID, list.OwnerID)
        if err != nil {
            return false, err
        }
        if isMember {
            return true, nil
        }
    }
    
    // Check list shares
    hasAccess, err := az.db.HasListShareAccess(userID, listID)
    if err != nil {
        return false, err
    }
    
    return hasAccess, nil
}

// Check if user can modify a list
func (az *AuthorizationService) CanModifyList(userID, listID string) (bool, error) {
    list, err := az.db.GetList(listID)
    if err != nil {
        return false, err
    }
    
    // Personal lists: only owner can modify
    if list.OwnerType == "user" {
        return list.OwnerID == userID, nil
    }
    
    // Tribe lists: any tribe member can modify
    if list.OwnerType == "tribe" {
        return az.db.IsUserTribeMember(userID, list.OwnerID)
    }
    
    return false, nil
}

// Check tribe membership
func (az *AuthorizationService) CanAccessTribe(userID, tribeID string) (bool, error) {
    return az.db.IsUserTribeMember(userID, tribeID)
}
```

### GraphQL Authorization

```go
// GraphQL resolver authorization example
func (r *listResolver) Items(ctx context.Context, obj *List) ([]*ListItem, error) {
    userID := auth.GetUserIDFromContext(ctx)
    
    // Check if user can access this list
    canAccess, err := r.authz.CanAccessList(userID, obj.ID)
    if err != nil {
        return nil, err
    }
    if !canAccess {
        return nil, errors.New("unauthorized access to list")
    }
    
    return r.db.GetListItems(obj.ID)
}

// Mutation authorization example
func (r *mutationResolver) AddListItem(ctx context.Context, listID string, input AddListItemInput) (*ListItem, error) {
    userID := auth.GetUserIDFromContext(ctx)
    
    // Check if user can modify this list
    canModify, err := r.authz.CanModifyList(userID, listID)
    if err != nil {
        return nil, err
    }
    if !canModify {
        return nil, errors.New("unauthorized to modify list")
    }
    
    return r.listService.AddItem(listID, input, userID)
}
```

## Frontend Authentication

### Token Storage

```typescript
// Auth utility functions
export class AuthService {
  private static readonly TOKEN_KEY = 'tribe_jwt_token';
  private static readonly USER_KEY = 'tribe_user';

  static setToken(token: string): void {
    try {
      localStorage.setItem(this.TOKEN_KEY, token);
    } catch (error) {
      // Fallback to sessionStorage if localStorage fails
      sessionStorage.setItem(this.TOKEN_KEY, token);
    }
  }

  static getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY) || 
           sessionStorage.getItem(this.TOKEN_KEY);
  }

  static removeToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
    sessionStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.USER_KEY);
    sessionStorage.removeItem(this.USER_KEY);
  }

  static isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }
}
```

### React Auth Context

```typescript
// Auth context for React
interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (email: string, name?: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const storedToken = AuthService.getToken();
    if (storedToken && !AuthService.isTokenExpired(storedToken)) {
      setToken(storedToken);
      // Fetch user info from API
      fetchUser(storedToken);
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = async (email: string, name?: string) => {
    setIsLoading(true);
    try {
      const response = await fetch('/api/v1/auth/dev/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, name })
      });
      
      const data = await response.json();
      
      AuthService.setToken(data.token);
      setToken(data.token);
      setUser(data.user);
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    AuthService.removeToken();
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{
      user,
      token,
      login,
      logout,
      isAuthenticated: !!token && !!user,
      isLoading
    }}>
      {children}
    </AuthContext.Provider>
  );
};
```

### GraphQL Client Auth

```typescript
// Apollo Client auth setup
import { setContext } from '@apollo/client/link/context';

const authLink = setContext((_, { headers }) => {
  const token = AuthService.getToken();
  
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    }
  }
});

const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache()
});
```

## Security Considerations

### Token Security
- Store JWT securely (httpOnly cookies in production)
- Implement token rotation for long-term sessions
- Use HTTPS everywhere
- Validate token on every request

### OAuth Security
- Validate state parameter to prevent CSRF
- Use secure redirect URLs
- Implement proper scope validation
- Store OAuth secrets securely

### Authorization Security
- Principle of least privilege
- Check permissions at multiple layers
- Log authorization failures
- Implement rate limiting

## Error Handling

### Authentication Errors

```go
// Common auth error types
var (
    ErrInvalidToken     = errors.New("invalid or expired token")
    ErrUnauthorized     = errors.New("unauthorized access")
    ErrForbidden        = errors.New("insufficient permissions")
    ErrOAuthFailed      = errors.New("OAuth authentication failed")
)

// Error response formatting
func (a *AuthService) HandleAuthError(c *gin.Context, err error) {
    switch err {
    case ErrInvalidToken:
        c.JSON(401, gin.H{
            "error": "invalid_token",
            "message": "Token is invalid or expired"
        })
    case ErrUnauthorized:
        c.JSON(401, gin.H{
            "error": "unauthorized",
            "message": "Authentication required"
        })
    case ErrForbidden:
        c.JSON(403, gin.H{
            "error": "forbidden",
            "message": "Insufficient permissions"
        })
    default:
        c.JSON(500, gin.H{
            "error": "auth_error",
            "message": "Authentication error occurred"
        })
    }
}
```

## Development Setup

### Environment Variables

```bash
# OAuth Configuration
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key
JWT_EXPIRY=168h  # 7 days

# Development
DEV_MODE=true
ALLOW_DEV_LOGIN=true
```

### Testing Authentication

```go
// Test helper for authentication
func createTestUser(t *testing.T) (*User, string) {
    user := &User{
        ID:       "test-user-id",
        Email:    "test@example.com",
        Name:     "Test User",
        Provider: "test",
    }
    
    authService := NewAuthService(testConfig)
    token, err := authService.generateJWT(user)
    require.NoError(t, err)
    
    return user, token
}

// Test auth middleware
func TestJWTAuthMiddleware(t *testing.T) {
    router := gin.New()
    authService := NewAuthService(testConfig)
    
    router.Use(authService.JWTAuthMiddleware())
    router.GET("/protected", func(c *gin.Context) {
        userID := c.GetString("user_id")
        c.JSON(200, gin.H{"user_id": userID})
    })
    
    user, token := createTestUser(t)
    
    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

---

*For API usage with authentication, see [API Design](./api-design.md)*
*For user-related database tables, see [Database Schema](./database-schema.md)* 