BEGIN;
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username TEXT UNIQUE NOT NULL,
                       password TEXT NOT NULL, -- для упрощения, в продакшн разработке пароли так не хранятся
                       created_at TIMESTAMP DEFAULT now()
);
CREATE TABLE tasks (
                       id SERIAL PRIMARY KEY,
                       user_id INT REFERENCES users(id) ON DELETE CASCADE,
                       title TEXT NOT NULL,
                       description TEXT,
                       status TEXT CHECK (status IN ('new', 'in_progress', 'done')) DEFAULT 'new',
                       created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_id ON tasks(id);
COMMIT;