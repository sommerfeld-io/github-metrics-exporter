# The AI assisted - or rather AI driven - journey from idea to v1.0.0

For quite some time I had the idea to take a look at some insights from my github actions workflows. Pipeline performance was interesting. but i did not really act on this any more than just installing the GitHub Integration into my Grafana CLoud instance. This already gives away that i rely on Grafana Cloud for my monitoring needs.

wanted info into pipeline performance, run time and see if it slows down over time, success/failure ratio and a single dashboard to see all info.

when thinking about this i quickly had lots of ideas i could implement into the exporter.

architecture foundation was clear very fast. implement in go because i always wanted to do a project with go for learning purposes. plus coincidentally go is the language that most exporters are written in. runtime should be docker. even though go could distribute a static binary, using docker makes it more versatile and easier to run in different environments and easy to update.  and i could align with my existing ways of shipping software through dockerhub. that way i could start with a pipeline setup i already had experience with and just needed to adopt it to the new project. alloy runs on all of my raspi nodes. one node shoiuld run the app as docker container and have alloy send data to grafana cloud (prometheus datasource).

that tells that my intention was to control the pipeline and SDLC myself. the AI should be the implementation helper for feature development. but the pipeline should enforce that this project sticks to solid and reproducible engineering principles. linting and automated testing was paramount.

i setup a github repo, places the initial project structure (linter, pipelines, folder structure, dev container, readme, license, etc.) myself and then just started writing github issues for my ideas. i had one big markdown file with unsorted notes and turned them with the help of ai into small tickets. these tickets were completely unsorted and needed refinement and prioritization and more information and lots of stuff. i had a workflow in place that turned them from "needs-triage" status into "review-me" status. an ai agent would then pick up the ticket, read the title, description and comments and generate a "real" description with requirements, acceptance criteria, implementation details etc. then i would review the ticket (this is the first human in the loop) and approve by moving from project column "skeleton" into "backlog".

> **TODO** link / explain / show the AI refinement workflow / Do some Diagramming here (no plantuml because github must render it)

next i sorted issues into milestones because lots of ideas meant i had different topics mixed around. i wanted to focus on metrics for Github actions workflows and Github actions workflows runs. tickets are sorted into milestone "1 - GitHub Actions Workflow Metrics" <https://github.com/sommerfeld-io/github-metrics-exporter/milestone/1>. this feature set should form version 1.0.0 when everything is implemented.

> **TODO** Then i started implementing with github copilot

## Implementation start and Claude Code vs GitHub Copilot

> **TODO** tbd ... what agent for what purpose ... how did i use them ...
>
> Almost right away, my monthly token budget was gone. plus at work i started getting familiar with claude code. so i though about switching to claude or using both.
>
> LINK TO ADR HERE !!!!!!

## Getting into a rhythm (= using the SDLC and pipeline to have Ai implement features)

> **TODO** ...
