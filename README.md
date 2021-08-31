# goerd

This tool allows you to turn schemas into instructions for the database, including migrations between schemas. Create easy-to-read data models as contracts for agreement between architects, development teams, and team leaders. This tool provides agility to change the huge data-layered models.

![Conceptual view](concept.png)

Features:

- Create posgresql migrations as a set of SQL queries that apply changes between two schemas, a schema and a database, or two databases using a schema definition that is stored in a yaml or plantuml file.
- Using https://github.com/jackc/pgx

Example of generated plantuml:

![Plantuml view](plantuml-example.png)

### API

See [docs](https://pkg.go.dev/github.com/covrom/goerd).

### testing and examples 
```docker-compose up``` and see `./output/schema.yaml`