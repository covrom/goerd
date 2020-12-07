DROP TABLE if exists accounts;

CREATE TABLE accounts (
	user_id serial NOT NULL,
	parent_id int4 NULL,
	username varchar(50) NOT NULL,
	"password" varchar(50) NOT NULL,
	email varchar(255) NOT NULL,
	created_on timestamp NOT NULL,
	last_login timestamp NULL,
	CONSTRAINT accounts_email_key UNIQUE (email),
	CONSTRAINT accounts_pkey PRIMARY KEY (user_id),
	CONSTRAINT accounts_username_key UNIQUE (username),
	CONSTRAINT accounts_userparent_key FOREIGN KEY (parent_id) REFERENCES accounts(user_id) on delete CASCADE
);