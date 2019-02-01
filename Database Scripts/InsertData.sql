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

INSERT INTO [SkillType]
(
    [Name]
)
VALUES
(
    'Back End Language'
),
(
    'Back End Framework'
),
(
    'Front End Language'
),
(
    'Front End Framework'
),
(
    'Database'
),
(
    'Hosting Platform'
),
(
    'Business Platform'
),
(
    'Source Control'
);

INSERT INTO [SkillName]
(
    [Name],
    [SkillTypeId],
    [IsEnabled]
)
SELECT 'C#', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Language'
UNION
SELECT 'Java', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Language'
UNION
SELECT 'Node', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Language'
UNION
SELECT 'MySQL', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Language'
UNION
SELECT 'Microsoft SQL', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Language'
UNION
SELECT 'SQLite', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Language'
UNION
SELECT 'Amazon Web Services', [Id], 1 FROM [SkillType] WHERE [Name] = 'Hosting Platform'
UNION
SELECT 'Azure', [Id], 1 FROM [SkillType] WHERE [Name] = 'Hosting Platform'
UNION
SELECT '.NET Framework', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Framework'
UNION
SELECT 'Entity Framework', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Framework'
UNION
SELECT 'ASP.NET MVC', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Framework'
UNION
SELECT 'ASP.NET WebAPI 2', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Framework'
UNION
SELECT 'ASP.NET WebForms', [Id], 1 FROM [SkillType] WHERE [Name] = 'Back End Framework'
UNION
SELECT 'React', [Id], 1 FROM [SkillType] WHERE [Name] = 'Front End Framework'
UNION
SELECT 'Angular', [Id], 1 FROM [SkillType] WHERE [Name] = 'Front End Framework'
UNION
SELECT 'JavaScript', [Id], 1 FROM [SkillType] WHERE [Name] = 'Front End Language'
UNION
SELECT 'Power BI', [Id], 1 FROM [SkillType] WHERE [Name] = 'Business Platform'
UNION
SELECT 'Team Foundation Server', [Id], 1 FROM [SkillType] WHERE [Name] = 'Source Control'
UNION
SELECT 'Git', [Id], 1 FROM [SkillType] WHERE [Name] = 'Source Control'
--UNION
--SELECT 'SVN', [Id], 1 FROM [SkillType] WHERE [Name] = 'Source Control'
--UNION
--SELECT 'Mercurial', [Id], 1 FROM [SkillType] WHERE [Name] = 'Source Control'

INSERT INTO [SkillWordAlias]
(
    [SkillNameId],
    [Alias]
)
SELECT [Id], 'C Sharp' FROM [SkillName] WHERE [Name] = 'C#'
UNION
SELECT [Id], 'CS' FROM [SkillName] WHERE [Name] = 'C#'
UNION
SELECT [Id], 'Node js' FROM [SkillName] WHERE [Name] = 'Node'
UNION
SELECT [Id], 'Node.js' FROM [SkillName] WHERE [Name] = 'Node'
UNION
SELECT [Id], 'NodeJS' FROM [SkillName] WHERE [Name] = 'Node'
UNION
SELECT [Id], 'My SQL' FROM [SkillName] WHERE [Name] = 'MySQL'
UNION
SELECT [Id], 'MSSQL' FROM [SkillName] WHERE [Name] = 'Microsoft SQL'
UNION
SELECT [Id], 'T-SQL' FROM [SkillName] WHERE [Name] = 'Microsoft SQL'
UNION
SELECT [Id], 'MS SQL' FROM [SkillName] WHERE [Name] = 'Microsoft SQL'
UNION
SELECT [Id], 'AWS' FROM [SkillName] WHERE [Name] = 'Amazon Web Services'
UNION
SELECT [Id], '.NET' FROM [SkillName] WHERE [Name] = '.NET Framework'
UNION
SELECT [Id], 'NET' FROM [SkillName] WHERE [Name] = '.NET Framework'
UNION
SELECT [Id], 'EF' FROM [SkillName] WHERE [Name] = 'Entity Framework'
UNION
SELECT [Id], 'EntityFramework' FROM [SkillName] WHERE [Name] = 'Entity Framework'
UNION
SELECT [Id], 'MVC' FROM [SkillName] WHERE [Name] = 'ASP.NET MVC'
UNION
SELECT [Id], 'ASP.NET WebAPI' FROM [SkillName] WHERE [Name] = 'ASP.NET WebAPI 2'
UNION
SELECT [Id], 'WebForms' FROM [SkillName] WHERE [Name] = 'ASP.NET WebForms'
UNION
SELECT [Id], 'React Js' FROM [SkillName] WHERE [Name] = 'React'
UNION
SELECT [Id], 'ReactJs' FROM [SkillName] WHERE [Name] = 'React'
UNION
SELECT [Id], 'Angular Js' FROM [SkillName] WHERE [Name] = 'Angular'
UNION
SELECT [Id], 'AngularJs' FROM [SkillName] WHERE [Name] = 'Angular'
UNION
SELECT [Id], 'JS' FROM [SkillName] WHERE [Name] = 'JavaScript'
UNION
SELECT [Id], 'Java Script' FROM [SkillName] WHERE [Name] = 'JavaScript'
UNION
SELECT [Id], 'PowerBI' FROM [SkillName] WHERE [Name] = 'Power BI'
UNION
SELECT [Id], 'BI' FROM [SkillName] WHERE [Name] = 'Power BI'
UNION
SELECT [Id], 'TFS' FROM [SkillName] WHERE [Name] = 'Team Foundation Server'
UNION
SELECT [Id], 'Git Hub' FROM [SkillName] WHERE [Name] = 'Git'
UNION
SELECT [Id], 'GitHub' FROM [SkillName] WHERE [Name] = 'Git';