CREATE TABLE IF NOT EXISTS snippets (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    alias       TEXT     UNIQUE NOT NULL,
    content     TEXT     NOT NULL,
    language    TEXT     DEFAULT '',
    tags        TEXT     DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts
    USING fts5(alias, content, tags, content=snippets, content_rowid=id);

CREATE TRIGGER IF NOT EXISTS snippets_ai AFTER INSERT ON snippets BEGIN
    INSERT INTO snippets_fts(rowid, alias, content, tags)
    VALUES (new.id, new.alias, new.content, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS snippets_au AFTER UPDATE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, alias, content, tags)
    VALUES ('delete', old.id, old.alias, old.content, old.tags);
    INSERT INTO snippets_fts(rowid, alias, content, tags)
    VALUES (new.id, new.alias, new.content, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS snippets_ad AFTER DELETE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, alias, content, tags)
    VALUES ('delete', old.id, old.alias, old.content, old.tags);
END;
