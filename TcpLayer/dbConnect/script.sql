create table users (
    id varchar(100) binary primary key,
    userName varchar(100),
    nickName varchar(100),
    pass varchar(100),
    profilePicAddress varchar(100) 
);

LOAD DATA LOCAL INFILE '/Users/pengyu.zhao/Desktop/EntryTask/GoTest/employee.csv' 
INTO TABLE users
FIELDS TERMINATED BY ',' 
LINES TERMINATED BY '\n';

insert into users values (1,'TEST','TEST','TEST',1);