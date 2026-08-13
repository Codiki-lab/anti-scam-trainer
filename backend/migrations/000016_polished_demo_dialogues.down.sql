UPDATE chat_options o
SET option_text=s.option_text,
    counterparty_reaction=s.counterparty_reaction
FROM migration_000016_dialogue_snapshot s
WHERE s.id=o.id;

DROP TABLE migration_000016_dialogue_snapshot;
