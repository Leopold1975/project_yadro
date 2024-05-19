CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username varchar(32) UNIQUE, 
    passwordHash text,
    role varchar(32)
);

INSERT INTO users(username, passwordHash, role) 
VALUES 
('admin', '$2a$10$VZuEWfPuLvZZaG1tU9HJT.YFbLzCcnFVMzyxAR28Rfma9.8CbXgay', 'admin'),
('user1', '$2a$10$D73FEMgwSFQSX2z3WsmLfu1oKGPYQWXpRHZLGpYNU7o5qKASRIz/u', 'user'),
('user2', '$2a$10$M6eSIfnQ0M3vJPfOuIIvIu.ulqclNrDsMeBmnIaeH7szEDkYQTMnG', 'user')
ON CONFLICT(username) DO NOTHING;