# TaskFolio - Getting Started Guide

Welcome to TaskFolio! This guide will help you get started with the project, whether you're a new developer joining the team or a contributor.

## Table of Contents

1. [Understanding the Project](#understanding-the-project)
2. [Project Documentation](#project-documentation)
3. [Next Steps](#next-steps)
4. [Development Process](#development-process)
5. [Quick Reference](#quick-reference)

---

## Understanding the Project

### What is TaskFolio?

TaskFolio is a student task management system designed for Zip Code Wilmington. It helps students:
- Organize and track assignments and projects
- Collaborate with classmates on team tasks
- Monitor progress and productivity
- Stay on top of deadlines

### Project Goals

- **Simplicity**: Easy to use and intuitive interface
- **Performance**: Fast response times and real-time updates
- **Collaboration**: Seamless team work and communication
- **Reliability**: Stable, secure, and available 24/7

---

## Project Documentation

We have created comprehensive planning documents to guide the development of TaskFolio. Here's what each document contains:

### 📋 [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md)
**Your roadmap for building the app**

This is the master plan divided into 7 phases:
- Phase 1: Project Setup & Infrastructure
- Phase 2: Core Backend Development
- Phase 3: Frontend Development
- Phase 4: Integration & Features
- Phase 5: Testing & Quality Assurance
- Phase 6: Deployment & Documentation
- Phase 7: Post-Launch & Maintenance

**When to use**: Start here for the big picture of what needs to be built and in what order.

### 🗄️ [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)
**Complete database design**

Contains:
- Entity Relationship Diagram (ERD)
- All table definitions with SQL
- Indexes and constraints
- Common queries and optimizations
- Migration strategy

**When to use**: Reference when implementing data models, writing queries, or creating migrations.

### 🔌 [API_SPECIFICATION.md](./API_SPECIFICATION.md)
**REST API documentation**

Includes:
- All API endpoints (authentication, users, tasks, teams, etc.)
- Request/response formats
- Error handling
- Authentication flows
- WebSocket events

**When to use**: When implementing API endpoints or integrating frontend with backend.

### 📁 [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)
**Code organization and conventions**

Covers:
- Directory structure for backend (Go)
- Directory structure for frontend (React/Vue)
- File naming conventions
- Design patterns to use
- Environment variables
- Build commands (Makefile)

**When to use**: When creating new files, organizing code, or understanding project layout.

### 📝 [TECHNICAL_REQUIREMENTS.md](./TECHNICAL_REQUIREMENTS.md)
**What the system must do and how well**

Details:
- Functional requirements (features)
- Non-functional requirements (performance, security, scalability)
- Technical specifications
- Constraints and assumptions
- Success criteria

**When to use**: Understanding requirements before implementing a feature or during code reviews.

### 🗓️ [ROADMAP.md](./ROADMAP.md)
**Timeline and milestones**

Provides:
- Visual timeline of development phases
- Milestone definitions
- Resource allocation
- Risk mitigation plans
- Success metrics for each phase

**When to use**: Planning sprints, tracking progress, or estimating timelines.

### 🏃 [SPRINT_PLANNING.md](./SPRINT_PLANNING.md)
**Sprint planning template and example**

Contains:
- Sprint planning template
- Example Sprint 1 with stories and tasks
- Definition of Done
- Sprint ceremonies
- Team agreements

**When to use**: Planning sprints, breaking down stories, or running sprint ceremonies.

---

## Next Steps

### For New Developers

If you're joining the TaskFolio project, here's what to do:

#### Week 1: Understanding
1. ✅ Read this GETTING_STARTED.md guide
2. ✅ Read the [README.md](./README.md) for project overview
3. ✅ Review [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - understand all phases
4. ✅ Study [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) - understand data model
5. ✅ Review [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) - understand code organization

#### Week 1-2: Setup
6. ⏳ Set up your development environment:
   - Install Go 1.21+
   - Install PostgreSQL 15+
   - Install Node.js 18+
   - Install Docker Desktop
   - Set up your IDE (VS Code, GoLand, etc.)
7. ⏳ Clone the repository
8. ⏳ Get the project running locally
9. ⏳ Run existing tests (once available)

#### Week 2: First Contribution
10. ⏳ Pick a small task from the current sprint
11. ⏳ Create a feature branch
12. ⏳ Implement the feature
13. ⏳ Write tests
14. ⏳ Submit a pull request
15. ⏳ Participate in code review

### For Project Leads

If you're leading the project:

1. ✅ Review all planning documents
2. ⏳ Form the development team
3. ⏳ Set up communication channels (Slack/Discord)
4. ⏳ Schedule sprint planning meetings
5. ⏳ Begin Phase 1 implementation
6. ⏳ Track progress against the roadmap

### For Contributors

If you're contributing to TaskFolio:

1. ✅ Read the [README.md](./README.md)
2. ✅ Review [TECHNICAL_REQUIREMENTS.md](./TECHNICAL_REQUIREMENTS.md)
3. ⏳ Look for issues labeled "good first issue"
4. ⏳ Fork the repository
5. ⏳ Make your changes
6. ⏳ Submit a pull request

---

## Development Process

### Agile Methodology

We follow Agile/Scrum practices:

- **Sprint Duration**: 2 weeks
- **Daily Standups**: 15 minutes, every day
- **Sprint Planning**: 2 hours, start of sprint
- **Sprint Review**: 1 hour, end of sprint
- **Sprint Retrospective**: 1 hour, after review

### Workflow

```
┌─────────────────┐
│  Sprint Backlog │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Create Branch  │
│  feature/xyz    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Implement      │
│  Feature        │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Write Tests    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Code Review    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Merge to Main  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Deploy         │
└─────────────────┘
```

### Git Workflow

1. **Branch Naming**:
   - Features: `feature/[story-number]-[description]`
   - Bugs: `bugfix/[issue-number]-[description]`
   - Hotfixes: `hotfix/[description]`

2. **Commit Messages**:
   ```
   type(scope): subject

   body

   footer
   ```
   
   Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

3. **Pull Requests**:
   - Link to the story/issue
   - Describe what changed and why
   - Add screenshots for UI changes
   - Ensure CI passes
   - Request review from team members

### Code Review Guidelines

**Reviewers should check**:
- [ ] Code follows project conventions
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Performance is acceptable
- [ ] Code is readable and maintainable

**Review timeframe**: Within 24 hours

---

## Quick Reference

### Essential Commands (Once Project is Set Up)

#### Backend (Go)
```bash
# Run the API server
go run cmd/api/main.go

# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run linter
golangci-lint run

# Build
go build -o bin/api cmd/api/main.go

# Run migrations
go run cmd/migrate/main.go up
```

#### Frontend
```bash
# Install dependencies
npm install

# Run dev server
npm run dev

# Build for production
npm run build

# Run tests
npm test

# Run E2E tests
npm run test:e2e
```

#### Docker
```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Rebuild containers
docker-compose build
```

### Important Links

- **Repository**: https://github.com/kristofer/TaskFolio
- **Issue Tracker**: GitHub Issues
- **Documentation**: This repository
- **CI/CD**: GitHub Actions

### Key Contacts

- **Project Lead**: [To be assigned]
- **Tech Lead**: [To be assigned]
- **Backend Lead**: [To be assigned]
- **Frontend Lead**: [To be assigned]

---

## Learning Resources

### Go Programming
- [Official Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### PostgreSQL
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [PostgreSQL Tutorial](https://www.postgresqltutorial.com/)

### React (if chosen)
- [React Documentation](https://react.dev/)
- [React Tutorial](https://react.dev/learn)

### Vue.js (if chosen)
- [Vue.js Documentation](https://vuejs.org/)
- [Vue.js Guide](https://vuejs.org/guide/)

### Docker
- [Docker Documentation](https://docs.docker.com/)
- [Docker Getting Started](https://docs.docker.com/get-started/)

---

## Frequently Asked Questions

### Q: Where do I start?
**A**: Read the README.md and IMPLEMENTATION_PLAN.md first. Then set up your development environment and look for a task in the current sprint.

### Q: How do I know what to work on?
**A**: Check the sprint backlog in the project management tool or GitHub Issues. Look for unassigned tasks or ask the tech lead.

### Q: What if I find a bug?
**A**: Create an issue in GitHub with details about how to reproduce it. If it's critical, notify the team immediately.

### Q: How do I ask for help?
**A**: Post in the team chat channel or ask during daily standup. Don't hesitate - we're all here to help!

### Q: Can I suggest a new feature?
**A**: Absolutely! Create a feature request issue or bring it up during sprint planning.

### Q: How often should I commit code?
**A**: Commit frequently (multiple times per day). Each commit should be a logical unit of work.

### Q: When should I create a pull request?
**A**: Create a PR when your feature is complete, tested, and ready for review. Don't wait too long!

### Q: How do I handle merge conflicts?
**A**: Pull the latest changes from main, resolve conflicts locally, test thoroughly, then push again.

---

## Tips for Success

1. **Communicate Early and Often**: Don't wait until you're stuck. Ask questions!

2. **Read the Code**: The best way to learn the codebase is to read it. Start with simple files and work your way up.

3. **Write Tests**: Don't skip tests. They save time in the long run and give you confidence in your code.

4. **Keep PRs Small**: Smaller PRs are easier to review and less likely to have issues.

5. **Document as You Go**: Update documentation when you change things. Future you (and others) will thank you.

6. **Follow Conventions**: Consistency makes the codebase easier to navigate and maintain.

7. **Learn from Reviews**: Code reviews are learning opportunities, not criticisms.

8. **Stay Updated**: Pull from main regularly to avoid conflicts and stay current with the latest changes.

---

## Welcome to the Team!

We're excited to have you working on TaskFolio! This is a collaborative project, and every contribution matters. Don't hesitate to ask questions, suggest improvements, or share ideas.

Let's build something great together! 🚀

---

**Last Updated**: October 16, 2025  
**Version**: 1.0  
**Maintained by**: TaskFolio Team
