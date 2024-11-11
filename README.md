# Gophis Social

**Gophis Social** is a on-development social network written in Go to explore and consolidate multiple concepts to build a _reliable_, _maintainable_, and _scalable_ backend service.

It's started as a project to follows [SelfMadeEngineer](www.skool.com/self-made-engineer)'s Backend Course.

## Technologies

### Used

- Go 1.22
- Docker
- Postgres running on Docker
- Swagger for documentation. Which requires [swag](https://github.com/swaggo/swag).

## Others concepts applied

* Twelve factor.

## Possible futures changes

- [ ] Support HTTP 2
- [ ] Replace `chi` with own implementation.

## Planned changes

- [x] Use of indexes
- [ ] Change `is_deleted` to `deleted_at` on the database.
