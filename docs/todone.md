# TODOne 1Pager
Doc to scope TODOne functionality

## Overview
TODOne will aggregate todos from various sources, make a prioritized list of items, and book a session for the user-specified time including a subset of items that can reasonably be done. After TODOne has suggested a set of items, user can veto or reprioritize as they see fit.

## User Experience
Ideally, a TUI that a user can start with a specific profile (doc specifying what sources to aggregate). If we use a TUI, then agent can more naturally ask user when and how much time to book on the calendar. Without a TUI we can support an arg like `--when ""` so user can still provide natural language for when they want to schedule. 

## Sources
Places where TODOne will look for todos. The idea is the user doesn't _have_ to put everything in a work tracker/todolist for TODOne to have visbility. User can work where and how they work, and TODOne will aggregate items across all of their integrated services.

1. Work tracking service (ADO, Trello, etc)
1. Local repos 
  - `rg TODO`
1. Messaging app 

## TODO Processing
Once all todos are aggregated, agent processes items by assigning a priority and an estimated effort (in minutes)


## MVP

Will focus on core agent functionality and user customization.

### Sources
Cut work tracking integration, support local code repo, support a mocked messaging app.

For messaging app, user should be able to provide the name of people they want TODOne to aggregate tasks from. Getting a chat history will be a tool call and the service will just be mocked to provide sample chat data.


### Work Tracking
- CLI app that takes `--profile /path/to/profile.md --when "Tomorrow morning"`
- Mock messaging data for "Boss"
- Agent can use tool calling to get "Boss" msg history
- Agent can extract a todo from Boss's chat and report it to user
- Add sample project that has some TODO comment
- Agent can identify both Boss's TODO and code TODO and prioritize them
- Agent can say "I will schedule x hrs on <day> to review the following: <list of TODOs>"
