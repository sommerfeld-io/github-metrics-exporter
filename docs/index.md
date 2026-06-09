# Template Repository

[doc-website]: https://sommerfeld-io.github.io/template-repository
[github-repo]: https://github.com/sommerfeld-io/template-repository
[file-issues]: https://github.com/sommerfeld-io/template-repository/issues
[project-board]: https://github.com/orgs/sommerfeld-io/projects/1/views/1

This project is a template repository for creating new repositories with a pre-defined structure. This template repository is relevant for [sommerfeld.io](https://github.com/sommerfeld-io) projects.

!!! warning "Todo - Remove if not needed"
    Short introduction to the project, its goals, and objectives.

- [Documentation Website][doc-website]
- [Github Repository][github-repo]
- [Sonarcloud Code Quality and Security Analysis](https://sonarcloud.io/project/overview?id=sommerfeld-io_template-repository)
- [Where to file issues][file-issues]
- [Project Board for Issues and Pull Requests][project-board]

## Requirements and Features

This section outlines the basic requirements and features of the project. These requirements serve as a guideline for the development of the application and provide a clear overview of the expected functionality.

!!! warning "Todo - Remove if not needed"
    Short description of the functional requirements, driving forces, extract (or abstract) of requirements. Maybe link to GitHub issues (which are written as user stories) or to link some BDD-style executable specifications.

## Usage

!!! warning "Todo - Remove if not needed"
    The basic commands to get any response from the system. This could be a simple `docker run ...` command.

## Scope and Context

!!! warning "Todo - Remove if not needed"
    Who is the target audience of the software? Who are the stakeholders? Use a list when stakeholders are clearly identifiable. Otherwise write a short description (e.g. when stakeholders are possibly everybody because the software is offered as open source to the whole world). Scope and context delimits your system (i.e. your scope) from all its communication partners (neighboring systems and users, i.e. the context  of your system). It thereby specifies the external interfaces. If necessary, differentiate the business context (domain specific inputs and outputs) from the technical context (channels, protocols, hardware).

## Architecture Constraints

!!! warning "Todo - Remove if not needed"
    Any requirement that constraints software architects in their freedom of design and implementation decisions or decision about the development process. These constraints sometimes go beyond individual systems and are valid for whole organizations and companies. Only list constraints that are not covered in the [Development Guide](https://github.com/sommerfeld-io/.github/blob/main/docs/development-guide.md).

## Building Block View

!!! warning "Todo - Remove if not needed"
    The building block view shows the components of the system and their relationships. It is a high-level view of the system architecture. Maybe think of C4 model diagrams. Do not write too much text here.

```kroki-ditaa

+--------+    +--------+
|        |    |        |
|  Web   |    |  API   |
|        |    |        |
+--+-----+    +--+-----+
   |             |
   |             |
+--+-----+    +--+-----+
|        |    |        |
|  DB    |    |  DB    |
|        |    |        |
+--------+    +--------+
```

## Contact

Feel free to contact me via <sebastian@sommerfeld.io> or [raise an issue in this repository][file-issues].
