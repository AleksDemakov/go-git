create table objects(
    hash text primary key, 
    path text, 
    size real, 
    content text
    ) without rowid;
create table commits(
    hash text primary key, 
    author text, 
    email text,
    message text,
	date text
    ) without rowid;
create table commit_objects(
    commit_hash text, 
    object_hash text  
    );
create table _index(
    hash text primary key,
    path text,
    sze real,
    content text
    ) without rowid;
create table refs (
	name text,
	commit_hash text
);

insert into commits values (43, 'me', 'me@mail.com', 'mda', '2026-04-23 09:00:00');
insert into commits values (42, 'me', 'me@mail.com', 'mda', datetime('now'));

select hash from commits order by date desc limit 1 ;

