# Tasking - Student Manual

**Your Personal Task Management Guide for Zip Code Wilmington**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Understanding Kanban Boards](#understanding-kanban-boards)
4. [Your Dashboard](#your-dashboard)
5. [Adding Tasks](#adding-tasks)
6. [Task Categories Explained](#task-categories-explained)
7. [Organizing Your Work](#organizing-your-work)
8. [Working with Instructor-Assigned Tasks](#working-with-instructor-assigned-tasks)
9. [Instructor View Capabilities](#instructor-view-capabilities)
10. [Tasking vs GitHub Projects](#tasking-vs-github-projects)
11. [Tasking vs Trello](#tasking-vs-trello)
12. [Best Practices & Tips](#best-practices--tips)

---

## Introduction

**Welcome to Tasking!** This is your personal task management system designed specifically for Zip Code Wilmington coding bootcamp students. Tasking helps you stay organized, track your progress, and manage the many assignments, projects, and professional development activities you'll encounter during your journey.

### What is Tasking?

Tasking is a **multi-tenant web application** built specifically for bootcamp students and instructors. It provides:

- **Personal Task Management**: Create and track your own tasks
- **Instructor Assignments**: Receive and manage tasks assigned by instructors
- **Visual Organization**: Use a Kanban board to see your work at a glance
- **Progress Tracking**: Monitor your completion status and stay on schedule
- **Cohort Integration**: Organized by your cohort (e.g., Java11.2, Data6.2)

### Why Use Tasking?

During your bootcamp, you'll juggle:
- Individual programming labs
- Group programming projects
- Reading assignments
- Documentation tasks
- Resume building
- Professional development activities
- Demo presentations
- Research and design work

Tasking helps you **see everything in one place** and **stay in control** of your learning journey.

---

## Getting Started

### Accessing Tasking

1. **Open your web browser** and navigate to the Tasking application URL provided by your instructor
2. **Log in** with your assigned username and password
   - Your username is typically: `student1`, `student2`, etc. (or your assigned ID)
   - Your initial password will be provided by your instructor
3. **Dashboard**: After logging in, you'll see your personal Kanban board

### First-Time Setup

When you first log in:
1. Review any tasks already assigned to you by instructors
2. Familiarize yourself with the four status columns (explained below)
3. Consider creating a few personal tasks to get comfortable with the system

### Demo Account for Practice

If you're just exploring, you can use:
- **Username**: `student1`, `student2`, or `student3`
- **Password**: `password123`

**Note**: These are demo accounts. Your instructor will provide your real credentials.

---

## Understanding Kanban Boards

### What is a Kanban Board?

A **Kanban board** is a visual workflow management tool that originated in Japanese manufacturing and is now widely used in software development and personal productivity. The word "kanban" (看板) means "visual signal" or "card" in Japanese.

### Why Kanban Works

Kanban boards help you:
- **Visualize your work**: See all your tasks at once
- **Limit work in progress**: Focus on completing tasks before starting new ones
- **Improve flow**: Move tasks smoothly from start to completion
- **Identify bottlenecks**: Spot where work is getting stuck
- **Track progress**: See how much you've accomplished

### The Four Columns in Tasking

Your Tasking board has **four status columns**:

#### 1. **Pending** (To Do)
- Tasks you haven't started yet
- Newly assigned tasks appear here
- This is your backlog of work

**When to use**: This is where tasks start. Don't let this column get too overwhelming—prioritize what's most important.

#### 2. **In Progress** (Doing)
- Tasks you're actively working on
- Limited to what you can realistically handle at once

**When to use**: Move a task here when you actually start working on it. Try to limit how many tasks are "in progress" at once—focus is key!

#### 3. **Blocked** (Stuck)
- Tasks that are stopped due to dependencies or issues
- Need help from instructors, teammates, or external factors

**When to use**: If you can't proceed because you're waiting for something (feedback, help, access to resources), move it here so instructors can see you need assistance.

#### 4. **Completed** (Done)
- Tasks you've finished
- Your record of accomplishment

**When to use**: Celebrate! Move tasks here when they're fully complete. This column shows your progress over time.

### How Kanban Differs from a Simple To-Do List

**To-Do Lists**:
- Linear, one-dimensional
- Doesn't show work in progress
- Hard to see the big picture

**Kanban Boards**:
- Visual, spatial organization
- Shows work-in-progress clearly
- Provides workflow visibility
- Encourages finishing before starting

---

## Your Dashboard

When you log into Tasking, you'll see your **Kanban board** as the main interface.

### Dashboard Components

1. **Header**
   - Shows "Tasking" branding
   - Displays your username and role (Student)
   - Logout button

2. **Action Buttons** (above the board)
   - **Add Task**: Create a new personal task
   - **Groups** (if applicable): View your group assignments
   - **Upload Tasks**: Import multiple tasks from a CSV file (advanced)

3. **Kanban Board**
   - Four columns: Pending, In Progress, Blocked, Completed
   - Task cards in each column
   - Drag-and-drop capability to move tasks between columns

4. **Task Cards**
   - Display task title
   - Show due date (if set)
   - Display task type badge (color-coded)
   - Click to view full details

### Navigation Tips

- **Click on a task card** to view full details, edit, or delete
- **Drag tasks** between columns to update their status
- **Scroll within columns** if you have many tasks
- **Use the search/filter** (if available) to find specific tasks

---

## Adding Tasks

You can add two types of tasks:
1. **Personal tasks** you create for yourself
2. **Tasks assigned by instructors** (you cannot create these, only receive them)

### Creating a Personal Task

1. **Click the "Add Task" button** at the top of your dashboard
2. **Fill in the task form**:

   - **Title** (required): A short, descriptive name
     - Example: "Complete Lab 3 - Arrays"
   
   - **Description** (optional): Details about what needs to be done
     - Example: "Implement array sorting algorithms and complete all test cases"
   
   - **Due Date** (optional): When the task should be completed
     - Use the date picker or enter manually (YYYY-MM-DD format)
   
   - **Task Type** (optional): Select the category that best fits
     - See [Task Categories Explained](#task-categories-explained) below
   
   - **Resource URL** (optional): Link to relevant materials
     - Example: GitHub repository, Google Doc, reading material link

3. **Click "Add Task"** to save
4. **Your task appears** in the "Pending" column

### Quick Task Entry Tips

- **Keep titles short but descriptive**: "Study OOP Concepts" instead of "Study"
- **Add due dates** to stay accountable
- **Link resources** so you can quickly access materials
- **Choose the right task type** to organize your work visually

---

## Task Categories Explained

Tasking supports **nine task types** to help you categorize your work. Each type has a specific purpose and appears with a different colored badge for quick visual identification.

### 1. **Group Programming Project**
**What it is**: Collaborative coding projects where you work with other students

**When to use**: 
- Team-based programming assignments
- Collaborative application development
- Hackathon-style group projects

**Example**: "Build a full-stack web app with Team Alpha"

**Tips**: Link to your group's GitHub repository in the Resource URL field

---

### 2. **Individual Programming Lab**
**What it is**: Solo coding exercises and programming challenges

**When to use**:
- Daily coding labs
- Algorithm practice
- Individual coding assignments
- LeetCode-style challenges

**Example**: "Complete Lab 5: Linked Lists"

**Tips**: These are usually the most frequent tasks in your bootcamp

---

### 3. **Reading**
**What it is**: Reading assignments, documentation, or technical articles

**When to use**:
- Assigned textbook chapters
- Technical documentation review
- Blog posts or technical articles
- API documentation

**Example**: "Read 'Effective Java' Chapter 3"

**Tips**: Add the reading material link to Resource URL for quick access

---

### 4. **Documentation**
**What it is**: Writing documentation, README files, or technical writing

**When to use**:
- Creating README files
- Documenting your code
- Writing technical specifications
- Creating user guides

**Example**: "Write README for portfolio project"

**Tips**: Good documentation is as important as good code!

---

### 5. **Resume**
**What it is**: Building, updating, or reviewing your professional resume

**When to use**:
- Initial resume creation
- Resume updates after learning new skills
- Resume review sessions
- Portfolio building

**Example**: "Update resume with new Java skills"

**Tips**: Keep your resume updated throughout the bootcamp

---

### 6. **ProDev** (Professional Development)
**What it is**: Professional skills, networking, and career preparation activities

**When to use**:
- LinkedIn profile building
- Mock interviews
- Networking events
- Career coaching sessions
- Communication skills workshops

**Example**: "Complete LinkedIn profile optimization"

**Tips**: Professional development is just as important as technical skills

---

### 7. **Demo/Presentation**
**What it is**: Preparing and delivering presentations or project demos

**When to use**:
- Project demonstrations
- Presentation preparation
- Public speaking practice
- Showcase events

**Example**: "Prepare final project demo presentation"

**Tips**: Practice makes perfect—rehearse your demos!

---

### 8. **Research, Analysis or Design**
**What it is**: Research tasks, system design, or architectural planning

**When to use**:
- Technology research
- System architecture design
- Database schema planning
- UX/UI design work
- Competitive analysis

**Example**: "Research microservices architecture patterns"

**Tips**: Good design happens before coding starts

---

### 9. **Organize**
**What it is**: Planning, organizing, and meta-work about your work

**When to use**:
- Weekly planning
- Organizing your workspace
- Setting up development environment
- Creating study schedules

**Example**: "Plan sprint for week 6"

**Tips**: Spend time organizing to save time executing

---

### Using Task Types Effectively

**Color Coding**: Each task type appears with a different colored badge on your Kanban board, making it easy to see the mix of work you have.

**Filtering** (if available): Some views let you filter by task type to focus on specific work.

**Balance**: Aim for a balanced mix—you need coding practice, reading, professional development, and more!

---

## Organizing Your Work

Effective task management is a skill that will serve you beyond bootcamp. Here's how to use Tasking to stay organized.

### Daily Workflow

**Morning Review** (5 minutes):
1. Check your Pending column—what's due soon?
2. Review In Progress tasks—what can you finish today?
3. Check for new instructor-assigned tasks
4. Prioritize your top 2-3 tasks for the day

**During the Day**:
1. Move tasks to "In Progress" when you start them
2. Move tasks to "Blocked" if you get stuck and need help
3. Move tasks to "Completed" when finished (feels great!)
4. Add new tasks as they come up

**Evening Review** (5 minutes):
1. Update any task statuses
2. Add notes to tasks if needed
3. Plan for tomorrow

### Weekly Planning

**Sunday or Monday** (15 minutes):
1. Review all upcoming due dates
2. Create tasks for the week's known work
3. Break large projects into smaller tasks
4. Estimate your capacity realistically

### Prioritization Strategies

**1. Due Date Priority**:
- Start with tasks due soonest
- Leave buffer time for unexpected issues

**2. Dependency Priority**:
- Tasks that block other work come first
- Group projects often have dependencies

**3. Impact Priority**:
- High-impact tasks (projects, assessments) before low-impact tasks

**4. Energy Priority**:
- Do hard cognitive work (coding, design) when you're fresh
- Do lighter work (reading, organizing) when you're tired

### Managing Work in Progress

**Limit your WIP (Work in Progress)**:
- Try to keep no more than 2-3 tasks "In Progress" at once
- Finish tasks before starting new ones
- Context switching is expensive—focus pays off

**Why this matters**:
- Multiple partially-done tasks = nothing done
- Completed tasks = progress you can show
- Focus = better quality work

### Using the Blocked Column Effectively

**When to mark tasks as Blocked**:
- Waiting for instructor feedback
- Blocked by a group member's incomplete work
- Missing access to required resources
- Technical issues preventing progress
- Need clarification on requirements

**What to include**:
- Move the task to Blocked
- Add a comment/note explaining what's blocking you
- Notify your instructor if needed

**Benefits**:
- Instructors can see where students need help
- You can move on to other work
- Clear visibility of bottlenecks

### Breaking Down Large Tasks

Large assignments can be overwhelming. Break them into smaller tasks:

**Example**:
Instead of one task: "Complete Final Project"

Create multiple tasks:
- "Design database schema for final project"
- "Set up GitHub repo and project structure"
- "Implement user authentication"
- "Build REST API endpoints"
- "Create frontend components"
- "Write unit tests"
- "Deploy to production"

**Benefits**:
- Less overwhelming
- Clear next steps
- Visible progress
- Easier to track time

---

## Working with Instructor-Assigned Tasks

Instructors can assign tasks to you in several ways. Understanding how these work helps you stay organized.

### Types of Instructor Assignments

**1. Individual Assignment**:
- Assigned directly to you
- Appears in your Pending column
- You're the only student with this task

**2. Cohort Assignment**:
- Assigned to your entire cohort (e.g., all Java11.2 students)
- Everyone in your cohort sees the same task
- Common for lectures, readings, or assessments

**3. Group Assignment**:
- Assigned to a specific group of students
- All group members see the task
- Use for team projects

### Identifying Instructor-Assigned Tasks

Instructor-assigned tasks:
- You cannot edit the title or description
- You can update the status (move between columns)
- You can add your own notes/comments
- May have a "Created by: instructor1" indicator

### Your Responsibilities

**For instructor-assigned tasks, you can**:
- View task details
- Update the status (Pending → In Progress → Completed)
- Move to Blocked if you need help
- Add personal notes or comments
- View due dates and requirements

**You cannot**:
- Delete instructor-assigned tasks
- Change the title, description, or due date
- Modify task type or resource URL

### Managing Assigned Work

**Priority**: Instructor-assigned tasks are usually required coursework—prioritize them!

**Communication**: If an assigned task is unclear, mark it as Blocked and ask for clarification

**Completion**: Move assigned tasks to Completed when you've finished the work

---

## Instructor View Capabilities

It's important to understand what instructors can and cannot do with your board.

### What Instructors CAN Do

**View Your Progress**:
- Instructors can see your Kanban board
- They can see which tasks are Pending, In Progress, Blocked, or Completed
- They can see when you moved tasks to different statuses
- They can view your personal tasks

**Why this is helpful**:
- Instructors can identify students who need help
- They can see if you're falling behind
- They can offer support proactively
- They can track cohort-wide progress

### What Instructors CANNOT Do

**Modify Your Board**:
- Instructors **cannot** move your tasks between columns
- They **cannot** edit your personal tasks
- They **cannot** delete your tasks
- They **cannot** change your task statuses

**Why this matters**:
- Your board is yours to manage
- You control your workflow
- You decide when tasks move to Completed
- **You own your progress**

### Your Privacy

**What's visible**:
- Your task titles and descriptions
- Your task statuses and progress
- Tasks you've created
- How you're organizing your work

**What's private**:
- Your personal notes (if implemented)
- Your draft work
- Your actual code (unless linked via Resource URL)

### Working Together

This **view-only** approach means:
- You maintain ownership of your learning
- Instructors provide guidance, not micromanagement
- You develop self-management skills
- There's accountability without control

**If you need help**: Mark tasks as Blocked, and instructors will see you need support!

---

## Tasking vs GitHub Projects

Both Tasking and GitHub Projects use Kanban boards, but they serve different purposes. Here's how they compare:

### GitHub Projects

**What it is**: Project management tool integrated with GitHub repositories

**Primary purpose**: Managing software development within a repository

**Best for**:
- Tracking issues and pull requests
- Managing repository-specific work
- Team collaboration on a single codebase
- Open-source project management

**Key features**:
- Linked to issues and PRs
- Milestone tracking
- Repository-specific
- Public or private to repo
- Great for code-centric workflows

### Tasking

**What it is**: Personal task management system for bootcamp students

**Primary purpose**: Managing all aspects of your learning journey

**Best for**:
- Organizing work across multiple projects
- Tracking non-coding tasks (reading, professional development)
- Managing instructor assignments
- Personal productivity
- Cohort-wide organization

**Key features**:
- Multi-category task types
- Instructor assignment capability
- Cohort-based organization
- Personal and assigned work in one place
- Not tied to any specific repository

### When to Use Each

**Use GitHub Projects when**:
- Working on a specific repository
- Managing issues and pull requests
- Coordinating with team members on code
- You need integration with GitHub features

**Use Tasking when**:
- Managing your overall bootcamp workload
- Tracking reading assignments
- Managing professional development tasks
- Viewing instructor-assigned work
- Organizing tasks that don't live in a GitHub repo

### Using Both Together

**Many students use both!**

**GitHub Projects**: For your individual projects and group repositories
- Track bugs, features, and pull requests
- Manage project-specific work

**Tasking**: For your overall bootcamp organization
- See all your work in one place
- Track assignments across projects
- Manage non-coding tasks
- Include a link to your GitHub Project board in Tasking's Resource URL field!

**Example workflow**:
1. Create a Tasking task: "Complete authentication feature for Team Project"
2. Add Resource URL linking to your GitHub Project board
3. Track detailed issues in GitHub Projects
4. Update Tasking status when the overall feature is done

---

## Tasking vs Trello

Trello is a popular general-purpose Kanban tool. Here's how it compares to Tasking:

### Trello

**What it is**: General-purpose visual collaboration tool

**Primary purpose**: Flexible project management for any use case

**Best for**:
- General project management
- Team collaboration
- Personal productivity
- Any workflow you can imagine

**Key features**:
- Highly customizable boards
- Power-ups and integrations
- Labels, checklists, attachments
- Public and private boards
- Works for any industry or purpose
- Flexible column names and structures

**Strengths**:
- Extremely flexible
- Great integrations
- Beautiful UI
- Mobile apps
- Free tier available

**Limitations for bootcamp students**:
- No built-in instructor assignment features
- Requires manual setup for each cohort
- No predefined task categories for bootcamp work
- You need to create structure from scratch

### Tasking

**What it is**: Purpose-built task management for coding bootcamp students

**Primary purpose**: Managing bootcamp work and instructor assignments

**Best for**:
- Bootcamp-specific workflows
- Receiving instructor assignments
- Categorizing bootcamp work types
- Cohort-based organization

**Key features**:
- Predefined task categories for bootcamp work
- Instructor assignment capability
- Cohort integration
- Built specifically for student-instructor workflow
- Designed for Zip Code Wilmington context

**Strengths**:
- No setup needed—ready for bootcamp use
- Instructor can assign tasks directly
- Task types match bootcamp reality
- Integrated with your cohort
- View-only instructor access built in

**Limitations compared to Trello**:
- Less flexible (by design)
- Fewer integrations
- More opinionated structure
- Bootcamp-specific (not general-purpose)

### Key Differences

| Feature | Tasking | Trello |
|---------|---------|--------|
| **Purpose** | Bootcamp student task management | General-purpose project management |
| **Instructor Assignments** | ✅ Built-in | ❌ Manual workaround needed |
| **Task Categories** | ✅ Predefined for bootcamp | ⚠️ Manual labels |
| **Customization** | ⚠️ Limited (structured) | ✅ Highly flexible |
| **Cohort Integration** | ✅ Built-in | ❌ Manual setup |
| **Learning Curve** | ✅ Quick (opinionated design) | ⚠️ Requires setup |
| **Mobile App** | ❌ Web only (currently) | ✅ Native apps |
| **Integrations** | ⚠️ Limited | ✅ Extensive |
| **Cost** | ✅ Free for students | ⚠️ Free tier + paid plans |

### When to Use Each

**Use Tasking when**:
- You're a Zip Code Wilmington student
- You want instructor-assigned tasks
- You want a pre-structured bootcamp workflow
- You need to see all bootcamp work in one place

**Use Trello when**:
- You want to plan personal projects
- You need deep customization
- You want mobile apps
- You're organizing non-bootcamp life (personal tasks, hobbies, etc.)
- You want extensive integrations

**Use both**:
- Tasking for bootcamp work
- Trello for personal life, side projects, or extracurricular activities

### Why Tasking Exists

Tasking was built specifically because general tools like Trello don't have:
- **Instructor assignment features**: No built-in way for instructors to assign work
- **Bootcamp task types**: You'd need to create labels manually
- **Student-instructor workflow**: No view-only instructor access
- **Cohort organization**: No built-in multi-tenant structure

**Tasking = Trello + bootcamp-specific features - unnecessary flexibility**

---

## Best Practices & Tips

### Productivity Tips

**1. Start Each Day with Your Board**
- Make Tasking your browser homepage during bootcamp
- Review your board first thing in the morning
- Know your priorities before you start coding

**2. Keep Tasks Small**
- Tasks should take hours, not days
- If a task takes more than 4-6 hours, break it down
- Small tasks = more frequent completion = more motivation

**3. Use Due Dates Wisely**
- Add due dates to important tasks
- Leave buffer time before hard deadlines
- Use due dates to prioritize

**4. Limit Work in Progress**
- Finish tasks before starting new ones
- Multitasking is a myth
- Focus = faster completion

**5. Update Regularly**
- Update your board at least twice daily
- Don't let it get stale
- Accurate boards are useful boards

### Organization Tips

**1. Name Tasks Clearly**
- Use descriptive titles: "Complete Arrays Lab" not "Lab 3"
- Include context: "Study for Week 6 Assessment"
- Make it scannable

**2. Use Task Types**
- Assign the right category to every task
- Visual variety helps you see balance
- Mix of task types = holistic learning

**3. Link Resources**
- Always add Resource URLs when relevant
- Link to GitHub repos, Google Docs, readings
- Save time finding materials later

**4. Add Descriptions**
- Use the description field for requirements
- Note what "done" looks like
- Include acceptance criteria

### Study Tips

**1. Balance Your Work**
- Don't just do coding tasks
- Include reading, professional development, documentation
- Well-rounded = better prepared

**2. Use Blocked Status**
- Don't suffer in silence
- Mark tasks as Blocked when you need help
- Instructors can't help if they don't know

**3. Celebrate Completions**
- Moving tasks to Completed feels good!
- Track your progress
- Reflect on what you've accomplished

**4. Weekly Reviews**
- Every week, look at your Completed column
- Recognize your progress
- Adjust your approach for next week

### Avoiding Common Mistakes

**❌ Don't**: Create one giant task per project
**✅ Do**: Break projects into multiple smaller tasks

**❌ Don't**: Ignore instructor-assigned tasks
**✅ Do**: Prioritize assigned work—it's required!

**❌ Don't**: Keep all tasks in Pending forever
**✅ Do**: Move tasks through the workflow actively

**❌ Don't**: Have 10+ tasks In Progress
**✅ Do**: Limit WIP to 2-3 tasks maximum

**❌ Don't**: Forget to add due dates
**✅ Do**: Use due dates to stay accountable

**❌ Don't**: Work in isolation when blocked
**✅ Do**: Mark tasks as Blocked and ask for help

### Time Management

**Pomodoro Technique + Tasking**:
1. Pick a task (move to In Progress)
2. Work for 25 minutes (focused)
3. Take a 5-minute break
4. Repeat 4 times, then take a longer break
5. When done, move to Completed

**Time Boxing**:
- Allocate specific time blocks to tasks
- Add estimates in task descriptions
- Track how long tasks actually take
- Improve your planning over time

### Staying Motivated

**1. Visualize Progress**
- Your Completed column is your trophy case
- Watch it grow week by week
- Celebrate small wins

**2. Set Weekly Goals**
- "This week I'll complete 10 tasks"
- "This week I'll finish the authentication module"
- Goals + tracking = achievement

**3. Share Your Board**
- Discuss your progress with peers
- Learn from how others organize
- Accountability through visibility

**4. Reflect and Adapt**
- What worked this week?
- What didn't?
- Adjust your approach
- Continuous improvement

---

## Getting Help

### If You Have Questions

**About Tasking**:
- Ask your instructor
- Consult this manual
- Check with fellow students

**About Your Tasks**:
- Mark as Blocked if you're stuck
- Ask in your cohort chat
- Schedule time with your instructor

### Support Resources

- **Tasking User Guide**: This document
- **Instructor Office Hours**: For task clarification
- **Cohort Slack/Discord**: Peer support
- **Zip Code Resources**: General bootcamp support

---

## Conclusion

**Tasking is your command center** for navigating the bootcamp experience. By using it effectively, you'll:

✅ Stay organized across multiple projects and assignments
✅ Never miss important deadlines
✅ Communicate your progress clearly to instructors
✅ Develop professional task management skills
✅ Reduce stress by seeing the big picture
✅ Build habits that will serve you in your career

**Remember**: The best task management system is the one you actually use. Make Tasking part of your daily routine, and it will pay dividends throughout your bootcamp and beyond.

---

**Ready to get started?** Log in to Tasking and start organizing your bootcamp journey today!

**Questions?** Ask your instructor or check the [main documentation](./README.md).

**Good luck with your studies!** 🚀

---

*Document Version: 1.0*  
*Last Updated: February 2026*  
*For: Zip Code Wilmington Coding Bootcamp Students*
