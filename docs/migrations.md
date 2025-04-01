<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->
# Database Schema Migration

The application catalog uses ORM [entgo.io], which supports both automatic migration and versioned migration.

While automatic migration is easy, it is also limited to relatively simple schema changes. In order to provide
a more robust migration capabilities, the application catalog is switching to using the version migration approach, which
is facilitated by entgo ORM and [Atlas] - an open-source database schema management system.

Both entgo and Atlas work together to provide the tooling for projects to make versioned schema migrations 
relatively easy and streamlined. There are two high-level stages in this process:

1) Generating migrations
2) Applying migrations

The entgo ORM framework supports DB schema generation from Go code, which also allows it to generate 
change records that track how the schema changes and from that generated versioned migration files.

## Generating Migrations
There are several phases in this stage. Each is briefly described in the sections below in concrete terms of
how it applies to the catalog service project.

### Generating Schema, Go code, migration diffs, and migration generation code
When the developer makes changes to the `internal/ent/schema` package, they will need to run the following to regenerate
the schema:
```bash
make ent-generate
```

This step will now also result in creation of files under `internal/ent/migrate/migrions`
directory to track the changes and hashes of those changes. It will also create code in support of the migrations 
in the `internal/ent/generated/migrate` package. Just as before, all the generated artifacts should be checked-in.
Therefore, up to this point, there is no difference in the development process.

### Generating Migration Scripts/Code

In preparing for a release of the product, the team also needs to prepare the code to execute the schema migration,
coalescing possibly several independent changes to the schema made over time. This is accomplished using the following
command,  where the `migration-name` parameter gives the aggregated set of schema changes a name.
```bash
make migration-generate MIGRATION=<migration-name>
```

Running this command, will bring up a temporary Postgres database instance, which will be used to replay the
entire migration history to compute the current schema state and to produce a set of SQL commands to migrate the schema
to the new/future state. These commands, along with hashes will be stored in `internal/ent/migrate/migrations` directory,
effectively extending the schema evolution records and thus supporting future migrations. The generated migration
records as well as the generated SQL commands should be checked-in.

#### Validating generated scripts

The Atlas tooling also includes a linter which can provide validation of the generated migration scripts and
can serve as a means to alert the developer or DB admin about possible problems in the upcoming migration.
To run this validation, simply execute this command:
```bash
make migration-lint
```
The output of this command may include warning about potential issues, which may need to be addressed by
additional manually written code. If schema changes are made deliberately, this should not be necessary, but sometime
it may not be entirely avoidable.

At this point, the developer or DB admin have all the collateral to apply (or execute) the migration.

## Applying Migrations

_Note: This section hasn't yet been vetted as the application catalog service presently ignites the postgress
DB instance as part of a single Helm chart. To support versioned migration, we must first separate deployment
of the underlying DB instance from the deployment of the catalog (or other) services._

Before applying the migration on a production system, it is a good idea to first do a dry-run of the migration
scripts on an off-line copy of the production system to make sure there are no unexpected issues, which could
cause the migration to leave the production system in an unusable state.

The only difference between applying the migration to a "test" copy of the production system and the production
system itself is the URL of the DB instance and possibly also the access credentials. Otherwise, the process
is identical. The task is accomplished using the `atlas migrate apply` command:

```bash
atlas migrate apply \
  --dir "file://internal/ent/migrate/migrations" \
  --url "postgres://root:pass@localhost:5432/application-catalog"
```

_Note: The --url option specified above is just an example. Obviously, this would differ in a production environment`

## More documentation

More detailed documentation is provided on [entgo.io versioned migrations] page. The above page provides merely a summary and a distillation 
of concrete steps to be executed in the context of this specific project and database drivers.


[entgo.io]: https://entgo.io/docs/versioned-migrations/
[Atlas]: https://atlasgo.io/getting-started/
[entgo.io versioned migrations]: https://entgo.io/docs/versioned-migrations/
