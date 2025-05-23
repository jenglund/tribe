# Development Guidelines

**Related Documents:**
- [Testing Strategy](./testing-strategy.md) - Testing standards and practices
- [Implementation Roadmap](./roadmap.md) - Development phases and milestones
- [Architecture](../architecture.md) - System architecture principles

## Code Quality Standards

### Test-Driven Development (TDD)
- **Write tests before implementation** for all new features
- Follow the Red-Green-Refactor cycle:
  1. **Red**: Write a failing test that describes the desired behavior
  2. **Green**: Write minimal code to make the test pass
  3. **Refactor**: Improve code quality while keeping tests green
- **Coverage Requirements**:
  - Backend: 90% line coverage, 95% for critical business logic
  - Frontend: 85% line coverage, 90% for business components
  - E2E: Cover all primary user journeys

### Clear Documentation
- **Comment complex business logic thoroughly**, especially:
  - Decision-making algorithms
  - Filter logic
  - Database transaction boundaries
  - Authentication/authorization flows
- **Use descriptive function and variable names**
- **Document API endpoints** with clear examples
- **Keep README files updated** with setup and usage instructions

### Type Safety
- **Leverage TypeScript strictly** in frontend code:
  - Enable `strict` mode in TypeScript configuration
  - Use proper type definitions for all props and state
  - Avoid `any` types except in specific integration cases
- **Use Go's type system fully**:
  - Define clear struct interfaces for all data types
  - Use type assertions carefully with proper error handling
  - Leverage Go's interface system for dependency injection

### Error Handling
- **Implement comprehensive error handling** at all layers:
  - Database operations with transaction rollback
  - External API calls with retry logic
  - User input validation with clear error messages
  - Authentication failures with appropriate status codes
- **Use structured logging** for debugging and monitoring
- **Implement graceful degradation** where possible

### Consistent Patterns
- **Follow established architectural patterns**:
  - Repository pattern for data access
  - Service layer for business logic
  - Clear separation of concerns
  - Dependency injection for testability

## Development Workflow

### Incremental Changes
- **Make small, verifiable changes** that can be tested independently
- **Commit frequently** with descriptive commit messages
- **Use feature branches** for all development work
- **Submit pull requests** for code review before merging

### Test Coverage Requirements
- **Ensure new code has appropriate test coverage**:
  - Unit tests for all business logic
  - Integration tests for API endpoints
  - Component tests for React components
  - E2E tests for critical user journeys
- **Run tests locally** before submitting pull requests
- **Verify tests pass in CI/CD pipeline** before deployment

### Schema Consistency
- **Database changes must include proper migrations**:
  - Forward migrations for schema changes
  - Backward migrations for rollback capability
  - Test migrations on sample data
  - Document breaking changes clearly
- **Version control all schema changes**
- **Never modify existing migrations** once they're in production

### API Consistency
- **Follow GraphQL schema conventions**:
  - Use descriptive field names
  - Implement proper error handling
  - Provide comprehensive type definitions
  - Use fragments for reusable query parts
- **Follow REST conventions** for simple operations:
  - Use appropriate HTTP methods
  - Implement consistent error responses
  - Follow URL naming conventions
  - Provide clear documentation

### Documentation Updates
- **Update relevant documentation with changes**:
  - API documentation for new endpoints
  - User stories for feature changes
  - Database schema documentation
  - Implementation notes for complex features

## AI Agent Collaboration Guidelines

### Incremental Development
- **Start with the simplest working implementation**
- **Add complexity gradually** with each iteration
- **Verify functionality** at each step before proceeding
- **Document decisions and trade-offs** made during development

### Test-First Approach
- **Write failing tests** that describe the expected behavior
- **Implement minimal code** to make tests pass
- **Refactor for quality** while maintaining test coverage
- **Add integration and E2E tests** for complete feature validation

### Code Review Focus Areas
- **Business logic correctness**
- **Error handling completeness**
- **Performance considerations**
- **Security implications**
- **Test coverage adequacy**
- **Documentation clarity**

### Communication Patterns
- **Explain complex decisions** in code comments
- **Document API changes** clearly in pull requests
- **Highlight breaking changes** and migration requirements
- **Provide context** for architectural decisions

## Frontend Development Guidelines

### React Best Practices
- **Use functional components** with hooks
- **Implement proper error boundaries** for graceful error handling
- **Use React Context** for global state management
- **Avoid prop drilling** with appropriate state lifting
- **Implement proper cleanup** in useEffect hooks

### State Management
- **Use React Context + useReducer** for complex state
- **Keep component state local** when possible
- **Implement optimistic updates** with rollback capability
- **Cache API responses** appropriately to reduce network calls

### Performance Optimization
- **Use React.memo** for expensive components
- **Implement proper key props** for list items
- **Use lazy loading** for route-based code splitting
- **Optimize bundle size** with tree shaking and analysis

### UI/UX Standards
- **Follow mobile-first responsive design**
- **Implement proper loading states** with skeleton screens
- **Provide clear error feedback** to users
- **Ensure accessibility** with proper ARIA labels and semantic HTML
- **Use consistent design patterns** from shadcn/ui components

## Backend Development Guidelines

### Go Best Practices
- **Follow Go idioms** and conventions
- **Use interfaces** for dependency injection and testing
- **Implement proper error handling** with descriptive error messages
- **Use context.Context** for request scoping and cancellation
- **Follow package organization** conventions

### Database Interactions
- **Use prepared statements** to prevent SQL injection
- **Implement proper transaction management**
- **Use database connection pooling** efficiently
- **Implement proper indexing** for query performance
- **Use migrations** for all schema changes

### API Design
- **Implement consistent error responses**
- **Use proper HTTP status codes**
- **Implement rate limiting** for API protection
- **Use middleware** for cross-cutting concerns
- **Document all endpoints** clearly

### Security Considerations
- **Validate all user input** at the API boundary
- **Implement proper authentication** and authorization
- **Use HTTPS** for all communications
- **Sanitize data** before database operations
- **Log security events** for monitoring

## Testing Guidelines

### Unit Testing
- **Test business logic in isolation**
- **Use dependency injection** for testable code
- **Mock external dependencies** appropriately
- **Test edge cases and error conditions**
- **Maintain high test coverage** on critical paths

### Integration Testing
- **Test API endpoints** with real database interactions
- **Use test databases** for isolation
- **Test authentication and authorization** flows
- **Verify data persistence** and retrieval
- **Test error handling** at integration boundaries

### End-to-End Testing
- **Cover critical user journeys** completely
- **Test cross-browser compatibility**
- **Verify mobile responsiveness**
- **Test with realistic data** volumes
- **Include error scenarios** in test suites

### Performance Testing
- **Test API response times** under load
- **Monitor database query performance**
- **Test with large datasets** to identify bottlenecks
- **Measure frontend bundle sizes** and loading times
- **Profile memory usage** in long-running operations

## Security Guidelines

### Authentication & Authorization
- **Implement proper session management**
- **Use secure token storage** practices
- **Validate permissions** on every request
- **Implement proper logout** functionality
- **Monitor authentication events**

### Data Protection
- **Validate and sanitize** all user input
- **Use parameterized queries** for database operations
- **Implement proper data encryption** for sensitive information
- **Follow data retention** policies
- **Implement audit logging** for sensitive operations

### Infrastructure Security
- **Use environment variables** for sensitive configuration
- **Implement proper CORS** policies
- **Use security headers** appropriately
- **Keep dependencies updated** with security patches
- **Monitor security vulnerabilities** in dependencies

## Code Review Checklist

### Functionality
- [ ] Code implements requirements correctly
- [ ] Edge cases are handled appropriately
- [ ] Error handling is comprehensive
- [ ] Performance considerations are addressed

### Quality
- [ ] Code follows established patterns and conventions
- [ ] Tests provide adequate coverage
- [ ] Documentation is clear and helpful
- [ ] No obvious security vulnerabilities

### Integration
- [ ] API contracts are maintained
- [ ] Database migrations are proper
- [ ] Configuration changes are documented
- [ ] Deployment considerations are addressed

### Maintainability
- [ ] Code is readable and well-organized
- [ ] Complex logic is documented
- [ ] Dependencies are appropriate
- [ ] Technical debt is minimized

---

*These guidelines should be followed by all developers working on the Tribe application. They should be updated as the project evolves and new patterns emerge.* 