CREATE TABLE project_sections
(
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR NOT NULL,
    sort        INT     NOT NULL,
    description TEXT    NOT NULL
);

CREATE TABLE projects
(
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section     INTEGER NOT NULL REFERENCES project_sections (id),
    name        VARCHAR NOT NULL,
    icon        TEXT    NOT NULL,
    pinned      BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT    NOT NULL
);