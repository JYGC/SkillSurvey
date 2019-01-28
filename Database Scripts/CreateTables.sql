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

CREATE TABLE IF NOT EXISTS [SkillType]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [Name] TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS [SkillName]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [Name] TEXT UNIQUE,
    [SkillTypeId] INT,
    [IsEnabled] INT
);

CREATE TABLE IF NOT EXISTS [SkillWordAlias]
(
    [Id] INTEGER PRIMARY KEY ASC,
    [SkillNameId] INT,
    [Alias] TEXT UNIQUE
);