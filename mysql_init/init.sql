create database if not exists golang;
use golang;
create table if not exists `users`
(
    user_id int unsigned NOT NULL AUTO_INCREMENT,
    username varchar(100) UNIQUE,
    created_at DATETIME,
    primary key (user_id)
) ENGINE=InnoDB;

create table if not exists `chats`
(
    chat_id int unsigned NOT NULL AUTO_INCREMENT,
    chat_name varchar(100) UNIQUE,
    created_at DATETIME,
    primary key (chat_id)
) ENGINE=InnoDB;

create table if not exists `chats_users`
(
    id int unsigned NOT NULL AUTO_INCREMENT,
    chat_id int unsigned NOT NULL,
    user_id int unsigned NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

create table if not exists `messages`
(
    message_id int unsigned NOT NULL AUTO_INCREMENT,
    chat_id int unsigned NOT NULL,
    user_id int unsigned NOT NULL,
    text TEXT,
    created_at DATETIME,
    primary key (message_id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

