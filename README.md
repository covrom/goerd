# goerd [WIP]

This tool allows you to turn schemas into instructions for the database, including migrations between schemas. Create easy-to-read data models as contracts for agreement between architects, development teams, and team leaders. This tool provides agility to change the huge data model.

![Conceptual view](concept.png)

Features (in progress):

- Generating posgresql migrations as a set of SQL queries that apply changes between two schemas, a schema and a database, or two databases using a yaml schema definition
- Using https://github.com/jackc/pgx
- Run as grpc-microservice
- Use as library
- [Check rules for schema](https://wiki.postgresql.org/wiki/Don't_Do_This)
- [Generate CRUDs like postgrest](https://github.com/PostgREST/postgrest)

Example of generated plantuml:

![Plantuml view](plantuml-example.png)

### testing and examples 
```docker-compose up``` and see `./output/schema.yaml`