CREATE TABLE IF NOT EXISTS [JobPost]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [SiteId] INTEGER,
    [JobSiteNumber] TEXT,
    [Title] TEXT,
    [Body] TEXT,
    [PostedDate] TEXT,
    [City] TEXT,
    [Country] TEXT,
    [Suburb] TEXT,
    [CreateDate] DATETIME DEFAULT CURRENT_TIMESTAMP,
    [ProcessStatus] INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS [Site]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [Name] TEXT
);

CREATE TABLE IF NOT EXISTS [Word]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [Name] TEXT,
    [ClassifiedWordId] INT,
    [JobPostCreateDate] TEXT
);

CREATE TABLE IF NOT EXISTS [ClassifiedWord]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [Name] TEXT UNIQUE,
    [Type] INT
);

CREATE TABLE IF NOT EXISTS [ClassifiedWordAlias]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [ClassifiedWordId] INT,
    [Alias] TEXT UNIQUE
);

INSERT INTO [Site]
(
    [Name]
)
VALUES
(
    'Seek.com.au'
),
(
    'au.Jora.com'
);