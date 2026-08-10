UPDATE chat_options o SET option_text=s.option_text
FROM migration_000014_option_snapshot s WHERE s.id=o.id;

DROP TABLE migration_000014_option_snapshot;
