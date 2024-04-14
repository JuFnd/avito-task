-- Удаление внешних ключей
ALTER TABLE profile_role
    DROP CONSTRAINT IF EXISTS fk_profile,
    DROP CONSTRAINT IF EXISTS fk_role;

-- Удаление таблицы profile
DROP TABLE IF EXISTS profile;

-- Удаление таблицы role
DROP TABLE IF EXISTS role;

-- Удаление таблицы password
DROP TABLE IF EXISTS password;

-- Удаление таблицы profile_role
DROP TABLE IF EXISTS profile_role;

-- Создание таблицы password
CREATE TABLE password (
                          id SERIAL PRIMARY KEY,
                          value BYTEA
);

-- Создание таблицы profile_role
CREATE TABLE profile_role (
                              id SERIAL PRIMARY KEY,
                              profile_id INT,
                              role_id INT
);

-- Создание таблицы profile
CREATE TABLE profile (
                         id SERIAL PRIMARY KEY,
                         login TEXT NOT NULL UNIQUE,
                         password_id INT NOT NULL,
                         profile_role_id INT,
                         CONSTRAINT fk_password FOREIGN KEY (password_id) REFERENCES password (id),
                         CONSTRAINT fk_profile_role FOREIGN KEY (profile_role_id) REFERENCES profile_role (id)
);

-- Создание таблицы role
CREATE TABLE role (
                      id SERIAL PRIMARY KEY,
                      value TEXT
);

-- Добавление внешних ключей к таблице profile_role
ALTER TABLE profile_role
    ADD CONSTRAINT fk_profile FOREIGN KEY (profile_id) REFERENCES profile (id),
    ADD CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES role (id);

INSERT INTO role(value) VALUES ('user'), ('admin');