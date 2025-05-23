# Refined Prompt Guidance for Claude 4 Sonnet Thinking

## Core Development Principles

### Safety & Error Prevention
- **CRITICAL**: If you accidentally delete/destroy code, immediately alert me and stop trying to fix it yourself
- Always use exact line numbers when editing files to prevent accidental deletion
- Make smaller, targeted edits rather than large section replacements
- Verify file integrity after edits before proceeding

### Development Workflow
- **Test-driven approach**: Write tests first when possible, ensure high coverage (80%+)
- **Atomic changes**: Solve one issue at a time, verify it works, then iterate
- **Verification before deployment**: Always confirm tests pass before suggesting to run applications
- Avoid commands that hang indefinitely (servers, etc.) - defer these to me

### Context & Communication
- Use absolute paths when referencing project locations (all projects in `/Users/zeblith/reporoot`)
- Provide clear explanations of what you're doing and why
- If you encounter conflicting instructions or unclear requirements, highlight them and suggest resolutions
- Maintain TODO sections in README with "Known Issues" and "Future Work"

## What's Changed from Previous Guidance

Based on current best practices and Claude 4's capabilities, here's what I've streamlined:

### Removed/Simplified:
1. **Makefile paranoia** - You're smart enough to check what's actually in files before running commands
2. **Excessive line number emphasis** - Still important, but not the catastrophic failure mode it was
3. **Audio cue anxiety** - Less relevant with better session management
4. **Overly detailed debugging steps** - Your reasoning capabilities handle this naturally

### Enhanced/Kept:
1. **Safety-first file editing** - Still critical to prevent data loss
2. **Test-driven development** - Remains excellent practice
3. **Clear communication** - Essential for collaboration
4. **Documentation habits** - READMEs and TODO tracking very valuable

## Effective Prompting Techniques for You

### Structure Requests Clearly
- Start with clear objectives and constraints
- Provide relevant context (tech stack, patterns, existing code)
- Use numbered steps for complex multi-part tasks
- Specify desired output format explicitly

### Leverage Your Strengths
- **Planning first**: Ask you to analyze and plan before coding
- **Iterative refinement**: Use conversational turns to improve solutions
- **Context synthesis**: Provide multiple files/examples for pattern matching
- **Self-correction**: Point out issues and let you fix them

### Example of Good Prompting Style:
```
Task: Implement user authentication for our React app

Context:
- Using Redux for state management
- Express.js backend with JWT
- Following existing patterns in /auth directory

Requirements:
1. Login form component with validation
2. JWT storage and refresh logic
3. Protected route wrapper
4. Logout functionality

Please analyze the existing auth patterns first, then implement each component.
```

## Version Management Best Practices
(Keeping this as it's still valuable)

1. **Prefer upgrading over downgrading**
2. **Establish hierarchy**: Core infrastructure → libraries → utilities  
3. **Follow dependency chains logically**
4. **Accept breaking changes when necessary for better architecture**
5. **Document version decisions**

## Key Differences for Claude 4 Sonnet Thinking

1. **Extended reasoning**: You can think through complex problems more thoroughly
2. **Better context handling**: Less need to micromanage information flow
3. **Improved instruction following**: More reliable execution of complex requests
4. **Self-correction ability**: Better at catching and fixing your own mistakes

The key insight is to treat you more like a capable senior developer rather than a tool that needs constant supervision, while maintaining appropriate safety guardrails. 