# Markdown Component Examples

Material for MkDocs is packed with many great features that make technical writing a joyful activity. For details, see Material for MkDocs [Reference Documentation](https://squidfunk.github.io/mkdocs-material/reference).

## Admonitions

Admonitions, also known as call-outs, are an excellent choice for including side content without significantly interrupting the document flow. Material for MkDocs provides several different types of admonitions and allows for the inclusion and nesting of arbitrary content. See the [Material for MkDocs documentation](https://squidfunk.github.io/mkdocs-material/reference/admonitions) for details.

```markdown
!!! note
    Lorem ipsum dolor sit amet, consectetur adipiscing elit.
```

!!! note
    Lorem ipsum dolor sit amet, consectetur adipiscing elit.

---

```markdown
!!! note "Phasellus posuere in sem ut cursus"
    Lorem ipsum dolor sit amet, consectetur adipiscing elit.
```

!!! note "Phasellus posuere in sem ut cursus"
    Lorem ipsum dolor sit amet, consectetur adipiscing elit.

---

```markdown
!!! note ""
    Lorem ipsum dolor sit amet, consectetur adipiscing elit.
```

!!! note ""
    Lorem ipsum dolor sit amet, consectetur adipiscing elit.

Following is a list of type qualifiers provided by Material for MkDocs, whereas the default type, and thus fallback for unknown type qualifiers, is note:

- `note`
- `abstract`
- `info`
- `tip`
- `success`
- `question`
- `warning`
- `failure`
- `danger`
- `bug`
- `example`
- `quote`

## Content tabs

Sometimes, it's desirable to group alternative content under different tabs, e.g. when describing how to access an API from different languages or environments. Material for MkDocs allows for beautiful and functional tabs, grouping code blocks and other content.

```markdown
=== "Lorem ipsum"
    Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.

=== "At vero"
    At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
```

=== "Lorem ipsum"
    Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.

=== "At vero"
    At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

## Diagrams

To render diagrams, the documentation site the [mkdocs-kroki-plugin](https://pypi.org/project/mkdocs-kroki-plugin). This allows to render diagrams in various formats like PlantUML, Mermaid, Graphviz, and others (see [Kroki.io](https://kroki.io) for all supported diagram types). Use code-fences with a tag of kroki-`<Module>` to place the code with the wanted diagram inside.
