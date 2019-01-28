create extension if not exists citext;

create table users
(
	id serial not null constraint users_pkey primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone,
	public_id text not null constraint users_public_id_key unique,
	username citext not null unique,
	password text not null,
	email citext not null,
	first_name text not null,
	last_name text not null
);

create index idx_users_deleted_at
	on users (deleted_at);

create table chat_rooms
(
	id serial not null constraint chat_rooms_pkey primary key,
	public_id text unique not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone,
	name text not null constraint chat_rooms_name_key unique,
	owner_id integer not null references users(id)
);

create index idx_chat_rooms_deleted_at
	on chat_rooms (deleted_at);

create table chat_subscriptions
(
	id serial not null constraint chat_subscriptions_pkey primary key,
	created_at timestamp with time zone,
	updated_at timestamp with time zone,
	deleted_at timestamp with time zone,
	user_id integer not null references users(id),
	chat_room_id integer not null references chat_rooms(id)
);

create index idx_chat_subscriptions_deleted_at
	on chat_subscriptions (deleted_at);

create table chat_messages
(
	id serial not null constraint chat_messages_pkey primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone,
	public_id text not null,
	body text not null,
	user_id integer not null references users(id),
	chat_room_id integer not null references chat_rooms(id)
);

create index idx_chat_messages_deleted_at
	on chat_messages (deleted_at);
