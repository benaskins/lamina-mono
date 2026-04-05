# Migrate synd CLI to axon-base

**Area:** axon-synd
**Depends on:** axon-base

Wire the store once at the root command level via axon-base pool, share
across all subcommands. Replace `axon.OpenDB` / `axon.RunMigrations` with
`pool.NewPool` / `migration.Run`. Drop axon dependency for database handling.

This is the first consumer migration. Pattern established here applies to
axon-book, axon-task, and any future service.
