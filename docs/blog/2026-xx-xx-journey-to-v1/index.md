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

## Blog Post Memo: Orchestrating Claude Code and GitHub Copilot in Parallel

* The Genesis & Brainstorming Process
    * **The Problem Statement:** Wanted to use GitHub Copilot and Claude Code simultaneously on this repository without duplicating prompt instructions or causing AI conflicts. I want the github.com UI to work with copilot first but also (later) with claude code.
    * **AI as a Thought Partner:** Started the journey by brainstorming with an AI (gemini)to map out directory structures, identify edge cases (like cross-OS compatibility), and refine the architecture.
* Establishing "The Single Source of Truth"
    * **Moving Away from Fragmentation:** Why keeping separate `.github/copilot-instructions.md` and `claude.md` files fails (sync friction, conflicting rules, maintenance fatigue).
    * **The Solution:** Designing a central repository directory (`docs/instructions/`) to house a single, comprehensive `development-guidelines.md` (or "Project Bible").
    * **Architectural Separation:** Keeping the master document strictly **tool-agnostic** (focusing purely on project standards, architecture, and coding paradigms) and **out-of-scope for source code** (keeping raw source files out of the instructions to save token context).
* The Implementation Process (The Tech Stack Setup)
    * **Leveraging Native Linux Power:** For a user-only Linux environment, utilizing native symbolic links (`ln -s`) is the absolute simplest, zero-overhead way to keep tools aligned.
    * **Linking the Ecosystem:** Creating a symlink from the root or hidden configuration directories (like `.github/copilot-instructions.md`) directly back to the master file in `docs/instructions/`.
    * **Git and Cloud Readiness:** * Confirming that Git natively tracks symlinks seamlessly across machines.
    * Verifying that GitHub Actions (Ubuntu/macOS runners) natively respect and traverse these symlinks during CI/CD automated checks, ensuring the cloud environment sees the exact same rules.
* Key Takeaways for the Reader
    * Don't overcomplicate your automation; use basic OS primitives (like symlinks) before jumping to heavy scripts.
    * Treat AI instructions like code: DRY (Don't Repeat Yourself).
    * Bouncing ideas off an AI first can shave hours off workflow optimization and help catch environment-specific edge cases early.

## Getting into a rhythm (= using the SDLC and pipeline to have Ai implement features)

Implemented things very quickly with AI. The GitHub issues already existed in the project. Implementation speed is extremely high, especially for topics where you would first have to learn something yourself (a specific library, for example).

But you have to be careful how you use it. I quickly went off on unnecessary tangents and started optimizing logging and refactoring before I even had the first feature running. The config file was another example. Probably more useful than logging optimization, but since the tool was initially just for myself, hardcoded configuration would have been good enough at first.

So it is very easy to get sidetracked and lose sight of the actual goals because "just implementing this little fix/refactoring" is so easy and fast.

That raises the question of how much of a prototype you want versus whether you're already trying to build something more serious. From that perspective, the config file was probably a sensible decision after all and getting sidetracked to the config file was a good thing.

### after a few days if developing - the feeling of control and the need to keep an eye on the changeset

I get the feeling that i prefer claude code locally over github copilot agent on github.com because i still like running linters and tests manually after an ai implementation. this might be my personal issue of "letting go" because the AI runs tests and linters too. and the pipeline does too. so the safety-nets are in place. and on the PR i could see, if the ai made changes i don't like. still i feel more "in control" locally. this might change over time and maybe i have to force myself to use the copilot agent in github.com more to facilitate this "change over time".

also i feel that i have to keep a more disciplined eye on the changes made by the AI so i still align with DRY principle. as long as i do not introduce coupling with DRY, this is an important principle. but since the AI does the change anyway, it is easier to let it slip because the AI is the one who has to deal with this in the future. this however could very well turn out to be a very wrong assumption because if my app does not build someday in the future, this might be a very big mess I have to clean uo (or at least steer the AI into cleaning up). so better to avoid this mess to begin with and keep the AI on a tight leash. the ai should not be allowed to introduce technical debt just as human developers should not be allowed. Or rather not be allowed ti introduce and not clean up in the future because technical debt is a topic in itself. but with the speed of AI, i can have the AI clean this up right away. so i (hopefully) do not have as big a "need" to pile on technical debt.

also i practice inspecting tests, new tests and changes to tests which makes me feel more in control over the changeset. especially on the the acceptance tests keep a close eye.
