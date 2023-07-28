---
title: Adding a new user and database
group: PostgreSQL
---

```sql
create database mydb;
create user myuser with encrypted password 'mypass';
grant all privileges on database mydb to myuser;
```