# System Architecture

**Related Documents:**
- [Database Schema](./database-schema.md) - Detailed database design
- [API Design](./api-design.md) - API structure and endpoints
- [Authentication](./authentication.md) - Auth system details

## Technology Stack

### Core Technologies
- **Backend**: Go with Gin web framework
- **Frontend**: React with TypeScript
- **Database**: PostgreSQL 15+
- **API**: GraphQL for complex queries, REST for simple operations (hybrid approach)
- **Authentication**: OAuth 2.0 (Google primary) + JWT tokens
- **Deployment**: Docker-based for easy self-hosting

### Supporting Technologies
- **State Management**: React Context + useReducer (avoiding Redux complexity)
- **UI Framework**: shadcn/ui + Tailwind CSS
- **Testing**: 
  - Backend: Go testing package + testify
  - Frontend: Vitest + React Testing Library
  - E2E: Playwright
- **Build Tools**: Standard Go toolchain, Vite for frontend

## High-Level Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   React Frontend    │    │   Go Backend        │    │   PostgreSQL        │
│   (TypeScript)      │◄──►│   (Gin + GraphQL)   │◄──►│   (Database)        │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
         │                           │                          │
         │                           │                          │
         ▼                           ▼                          ▼
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   CDN/Static Files  │    │   External APIs     │    │   Redis (Cache)     │
│                     │    │   (OAuth, Maps)     │    │   (Optional)        │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

## Backend Structure

```
cmd/
├── server/
│   └── main.go                 # Application entry point
├── migrate/
│   └── main.go                 # Database migration tool
internal/
├── api/
│   ├── graphql/               # GraphQL schema and resolvers
│   ├── rest/                  # REST endpoints (auth, simple operations)
│   └── middleware/            # Auth, CORS, logging, rate limiting
├── auth/
│   ├── oauth.go              # OAuth providers
│   └── jwt.go                # JWT token management
├── models/                   # Domain models and validation
├── services/                 # Business logic layer
├── repository/               # Data access layer
├── filters/                  # Decision filtering engine
└── utils/                    # Shared utilities
migrations/                   # SQL migration files
tests/                       # Test files
docker/                      # Docker configuration
```

## Frontend Structure

```
src/
├── components/
│   ├── auth/                # Authentication components
│   ├── tribes/              # Tribe management
│   ├── lists/               # List management
│   ├── decisions/           # Decision-making flow
│   └── common/              # Reusable UI components
├── graphql/
│   ├── queries/             # GraphQL queries
│   ├── mutations/           # GraphQL mutations
│   └── fragments/           # Reusable fragments
├── hooks/                   # Custom React hooks
├── services/                # API client and utilities
├── store/                   # Context-based state management
├── types/                   # TypeScript definitions
└── utils/                   # Helper functions
```

## Design Principles

### Simplicity First
- Avoid over-engineering for MVP
- Choose proven technologies over cutting-edge
- Prioritize maintainability over performance optimization

### Self-Hosting Friendly
- Docker-based deployment
- Minimal external dependencies
- Clear setup documentation
- Environment-based configuration

### Scalability Considerations
- Database design supports growth
- Stateless backend design
- Caching strategy for high-traffic operations
- API design supports mobile clients

## Data Flow

### Authentication Flow
1. User initiates Google OAuth
2. Backend validates OAuth token
3. JWT issued for subsequent requests
4. Frontend stores JWT in localStorage
5. JWT included in GraphQL requests

### Decision Making Flow
1. User creates decision session
2. Lists and filters applied via GraphQL
3. Real-time elimination via WebSocket
4. Session state maintained in database
5. Final result logged to activity history

### List Management Flow
1. CRUD operations via GraphQL
2. Sharing permissions checked at API level
3. Activity logging for audit trail
4. Optimistic updates in frontend

## Security Architecture

### Authentication
- OAuth 2.0 for external identity
- JWT for session management
- Secure token storage practices

### Authorization
- Role-based access (tribe membership)
- Resource-level permissions
- API-level authorization checks

### Data Protection
- HTTPS everywhere
- Input validation and sanitization
- SQL injection prevention
- CORS policy enforcement

## Performance Considerations

### Database
- Proper indexing strategy (see [Database Schema](./database-schema.md))
- Connection pooling
- Query optimization
- Pagination for large datasets

### Caching Strategy
- Redis for session data (optional for MVP)
- Browser caching for static assets
- GraphQL query result caching

### Frontend Performance
- Code splitting for large bundles
- Lazy loading for non-critical features
- Optimistic UI updates
- Image optimization

## Deployment Architecture

### Development Environment
- Docker Compose for local development
- Hot reload for both frontend and backend
- Test database with sample data

### Production Environment
- Container orchestration (Docker Swarm or Kubernetes)
- Load balancer (nginx)
- Database backups and monitoring
- Log aggregation and monitoring

## Monitoring and Observability

### Logging
- Structured logging (JSON format)
- Request/response logging
- Error tracking and alerting

### Metrics
- API response times
- Database query performance
- User activity metrics
- System resource usage

### Health Checks
- Database connectivity
- External service availability
- API endpoint health

## Technology Rationale

### Go Backend
- Excellent performance for concurrent operations
- Strong typing and error handling
- Great ecosystem for web services
- Easy deployment (single binary)

### React Frontend
- Mature ecosystem and community
- Excellent TypeScript support
- Component reusability
- Strong testing tools

### PostgreSQL Database
- ACID compliance for data integrity
- JSON support for flexible data
- Excellent performance and scaling
- Strong backup and recovery tools

### GraphQL + REST Hybrid
- GraphQL for complex queries and real-time features
- REST for simple operations and file uploads
- Gradual migration path
- Familiar patterns for developers

---

*For implementation details, see [Implementation Roadmap](./implementation/roadmap.md)* 