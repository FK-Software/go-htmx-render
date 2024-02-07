# Go HTMX Render
An example of using Go + HTMX deployed on [Render](https://render.com)

Prerequisites to run the project:
- [Go 1.21](https://go.dev/dl) or later
- [PostgreSQL](https://postgresql.org)

### Creating the database
Run the following query in your database:
```sql
CREATE TABLE "tasks" (
    id serial NOT NULL PRIMARY KEY,
    title character varying(255) NOT NULL,
    created_at timestamp without time zone NULL,
    updated_at timestamp without time zone NULL
);
```

### Running the project
1. Clone the repo `git clone git@github.com:FK-Software/go-htmx-render`
2. Go to the repo's folder
3. Create an `.env` file like the following:
```dotenv
PORT="8080"
DATABASE_URL="postgres://user:pass@host/database?sslmode=disable"
```
4. Run `go mod tidy` to download the required Go modules
5. Run `go build` to build the binary
6. Execute the binary `ENV=dev ./go-htmx-render`

