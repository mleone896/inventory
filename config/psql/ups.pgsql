CREATE EXTENSION IF NOT EXISTS hstore;

CREATE TABLE IF NOT EXISTS colors (
		id serial,
		name varchar(256) not null,
		in_use bool not null default false,
		primary key (id),
		last_in_use timestamp without time zone default '2001-09-28 01:00:00',
		unique(name)
);
CREATE INDEX IF NOT EXISTS color_name_idx ON colors(name);

CREATE TABLE IF NOT EXISTS ec2_instances (
		id serial,
		instance_id varchar(256) not null,
		account_id varchar(256) not null,
    subnet_id varchar(256) not null,
		tags hstore,
		primary key (id),
		unique(instance_id, account_id)
);
CREATE INDEX IF NOT EXISTS instance_tags_idx ON ec2_instances(tags);


CREATE TABLE IF NOT EXISTS accounts (
		id serial,
		name varchar(256) not null,
		tags hstore,
		primary key (id),
		unique(name)
);

CREATE TABLE IF NOT EXISTS vpcs (
		id serial,
		vpc_id varchar(256) not null,
		account_id varchar(256) not null,
		tags hstore,
		primary key (id),
		unique(vpc_id, account_id)
);


CREATE TABLE IF NOT EXISTS subnets (
		id serial,
		vpc_id varchar(256) not null,
		subnet_id varchar(256) not null,
		availability_zone varchar(256) not null,
		account_id varchar(256) not null,
		tags hstore,
		primary key (id),
		unique(subnet_id, account_id)
);
CREATE INDEX IF NOT EXISTS subnet_id_idx ON subnets(subnet_id)



